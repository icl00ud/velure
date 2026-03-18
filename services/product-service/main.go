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
	if log == nil {
		log = logger.NewNop()
	}

	if err := deps.loadEnv(); err != nil {
		log.Info("No .env file found, using system environment variables")
	}

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
	handler := handlers.NewProductHandler(service)
	healthHandler := handlers.NewHealthHandler()

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
		AllowOrigins: resolveAllowedOrigins(),
		AllowHeaders: "Origin, Content-Type, Accept",
	}))
	app.Use(middleware.PrometheusMiddleware())

	api := app.Group("/")

	metricsHandler := promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
		},
	)
	app.Get("/metrics", adaptor.HTTPHandler(metricsHandler))

	api.Get("/health", healthHandler.Check)
	products := api.Group("/api/products")
	products.Get("", handler.GetProducts)
	products.Get("/categories", handler.GetCategories)
	products.Get("/count", handler.GetProductsCount)
	products.Post("", handler.CreateProduct)
	products.Patch("/:id/inventory", handler.PatchProductInventory)
	products.Put("/:id", handler.UpdateProduct)
	products.Delete("/:id", handler.DeleteProductById)
	products.Get("/:id", handler.GetProductById)

	return app
}

func resolveAllowedOrigins() string {
	const defaultAllowedOrigins = "https://velure.local"

	rawOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if rawOrigins == "" {
		return defaultAllowedOrigins
	}

	origins := strings.Split(rawOrigins, ",")
	filtered := make([]string, 0, len(origins))
	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}

	if len(filtered) == 0 {
		return defaultAllowedOrigins
	}

	return strings.Join(filtered, ",")
}

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
