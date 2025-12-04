package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"product-service/internal/config"
	"product-service/internal/handler"
	"product-service/internal/middleware"
	"product-service/internal/repository"
	"product-service/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/icl00ud/velure-shared/logger"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type appDependencies struct {
	loadEnv   func() error
	buildRepo func(cfg *config.Config) (repository.ProductRepository, func(context.Context) error, func(), error)
	newSvc    func(repository.ProductRepository) services.ProductService
	listen    func(app *fiber.App, addr string) error
}

var defaultDeps = appDependencies{
	loadEnv: func() error { return godotenv.Load() },
	buildRepo: func(cfg *config.Config) (repository.ProductRepository, func(context.Context) error, func(), error) {
		mongodb, err := config.NewMongoDB(cfg.MongoURI)
		if err != nil {
			return nil, nil, nil, err
		}

		redis := config.NewRedis(cfg.RedisAddr, cfg.RedisPassword)
		repo := repository.NewProductRepository(mongodb.Database(cfg.DatabaseName), redis)

		return repo, mongodb.Disconnect, func() {
			_ = redis.Close()
		}, nil
	},
	newSvc: services.NewProductService,
	listen: func(app *fiber.App, addr string) error {
		return app.Listen(addr)
	},
}

var log *logger.Logger

// fatalf is a variable to allow tests to replace the fatal behavior
var fatalf = func(v ...interface{}) {
	if log != nil {
		log.Fatal("fatal error", logger.Any("error", v))
	}
}

func main() {
	log = logger.Init(logger.Config{
		ServiceName: "product-service",
		Level:       os.Getenv("LOG_LEVEL"),
		UseColor:    os.Getenv("LOG_COLOR") != "false",
	})

	if err := run(defaultDeps); err != nil {
		fatalf("Failed to start server:", err)
	}
}

func run(deps appDependencies) error {
	// Ensure log is initialized
	if log == nil {
		log = logger.NewNop()
	}
	
	// Load environment variables from .env file if it exists
	if err := deps.loadEnv(); err != nil {
		log.Info("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.New()

	log.Info("Starting Product Service",
		logger.String("port", cfg.Port),
		logger.String("database", cfg.DatabaseName),
		logger.String("mongodb_uri", maskURI(cfg.MongoURI)),
		logger.String("redis_addr", cfg.RedisAddr))

	repo, mongoDisconnect, redisClose, err := deps.buildRepo(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize repositories: %w", err)
	}

	if mongoDisconnect != nil {
		defer func() {
			if err := mongoDisconnect(context.Background()); err != nil {
				log.Error("Error disconnecting from MongoDB", logger.Err(err))
			}
		}()
	}

	if redisClose != nil {
		defer redisClose()
	}

	// Initialize services
	service := deps.newSvc(repo)
	service.SyncProductCatalogMetric(context.Background())

	app := setupFiberApp(service)

	port := cfg.Port
	if port == "" {
		port = "3010"
	}

	log.Info("Product service started", logger.String("port", port))
	return deps.listen(app, ":"+port)
}

func setupFiberApp(service services.ProductService) *fiber.App {
	// Initialize handlers
	handler := handlers.NewProductHandler(service)
	healthHandler := handlers.NewHealthHandler()

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware - custom request logging
	app.Use(func(c *fiber.Ctx) error {
		path := c.Path()
		if path == "/metrics" || path == "/health" {
			return c.Next()
		}
		start := c.Context().Time()
		err := c.Next()
		log.Info("request",
			logger.String("method", c.Method()),
			logger.String("path", path),
			logger.Int("status", c.Response().StatusCode()),
			logger.String("duration", c.Context().Time().Sub(start).String()))
		return err
	})
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))
	app.Use(middleware.PrometheusMiddleware())

	// Routes
	api := app.Group("/")

	// Prometheus metrics endpoint with error handling
	metricsHandler := promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
		},
	)
	app.Get("/metrics", adaptor.HTTPHandler(metricsHandler))

	// Health routes
	api.Get("/health", healthHandler.Check)

	// Helper function to register product routes on a group
	registerProductRoutes := func(g fiber.Router) {
		g.Get("/", handler.GetAllProducts)
		g.Get("/products/search", handler.SearchProducts)
		g.Get("/products", handler.GetProductsREST)
		g.Get("/getProductsByName/:name", handler.GetProductsByName)
		g.Get("/getProductsByPage", handler.GetProductsByPage)
		g.Get("/getProductsByPageAndCategory", handler.GetProductsByPageAndCategory)
		g.Get("/getProductsCount", handler.GetProductsCount)
		g.Get("/categories", handler.GetCategories)
		g.Get("/:id", handler.GetProductById)
		g.Post("/", handler.CreateProduct)
		g.Post("/updateQuantity", handler.UpdateProductQuantity)
		g.Delete("/deleteProductsByName/:name", handler.DeleteProductsByName)
		g.Delete("/deleteProductById/:id", handler.DeleteProductById)
	}

	// Product routes - local dev with Caddy rewrite
	products := api.Group("/product")
	registerProductRoutes(products)

	// Kubernetes ALB routes (no path rewriting)
	apiProducts := api.Group("/api/product")
	registerProductRoutes(apiProducts)

	return app
}

// maskURI masks sensitive information in MongoDB URI for logging
func maskURI(uri string) string {
	if len(uri) > 20 && uri[:10] == "mongodb://" {
		if idx := strings.Index(uri, "://"); idx != -1 {
			remaining := uri[idx+3:]
			if atIdx := strings.Index(remaining, "@"); atIdx != -1 {
				userPass := remaining[:atIdx]
				if colonIdx := strings.Index(userPass, ":"); colonIdx != -1 {
					user := userPass[:colonIdx]
					host := remaining[atIdx:]
					return uri[:idx+3] + user + ":***" + host
				}
			}
		}
	}
	return uri
}
