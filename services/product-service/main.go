package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"product-service/internal/config"
	"product-service/internal/handlers"
	"product-service/internal/middleware"
	"product-service/internal/repository"
	"product-service/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
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

var fatalf = log.Fatal

func main() {
	if err := run(defaultDeps); err != nil {
		fatalf(err)
	}
}

func run(deps appDependencies) error {
	// Load environment variables from .env file if it exists
	if err := deps.loadEnv(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.New()

	log.Printf("Starting Product Service with configuration:")
	log.Printf("- Port: %s", cfg.Port)
	log.Printf("- Database: %s", cfg.DatabaseName)
	log.Printf("- MongoDB URI: %s", maskURI(cfg.MongoURI))
	log.Printf("- Redis Address: %s", cfg.RedisAddr)

	repo, mongoDisconnect, redisClose, err := deps.buildRepo(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize repositories: %w", err)
	}

	if mongoDisconnect != nil {
		defer func() {
			if err := mongoDisconnect(context.Background()); err != nil {
				log.Printf("Error disconnecting from MongoDB: %v", err)
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

	log.Printf("Product service is running on port %s", port)
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

	// Middleware
	app.Use(logger.New(logger.Config{
		Next: func(c *fiber.Ctx) bool {
			path := c.Path()
			return path == "/metrics" || path == "/health"
		},
	}))
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
