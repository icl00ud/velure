package main

import (
	"context"
	"log"
	"os"
	"strings"

	"product-service/internal/config"
	"product-service/internal/handlers"
	"product-service/internal/repository"
	"product-service/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.New()

	log.Printf("Starting Product Service with configuration:")
	log.Printf("- Port: %s", cfg.Port)
	log.Printf("- Database: %s", cfg.DatabaseName)
	log.Printf("- MongoDB URI: %s", maskURI(cfg.MongoURI))
	log.Printf("- Redis Address: %s", cfg.RedisAddr)

	// Initialize database connections
	mongodb, err := config.NewMongoDB(cfg.MongoURI)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer func() {
		if err := mongodb.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	redis := config.NewRedis(cfg.RedisAddr, cfg.RedisPassword)
	defer redis.Close()

	// Initialize repository
	repo := repository.NewProductRepository(mongodb.Database(cfg.DatabaseName), redis)

	// Initialize services
	service := services.NewProductService(repo)

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
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Routes
	api := app.Group("/")

	// Health routes
	api.Get("/health", healthHandler.Check)

	// Product routes
	products := api.Group("/product")
	products.Get("/", handler.GetAllProducts)
	products.Get("/getProductsByName/:name", handler.GetProductsByName)
	products.Get("/getProductsByPage", handler.GetProductsByPage)
	products.Get("/getProductsByPageAndCategory", handler.GetProductsByPageAndCategory)
	products.Get("/getProductsCount", handler.GetProductsCount)
	products.Get("/categories", handler.GetCategories)
	products.Post("/", handler.CreateProduct)
	products.Delete("/deleteProductsByName/:name", handler.DeleteProductsByName)
	products.Delete("/deleteProductById/:id", handler.DeleteProductById)

	port := os.Getenv("PRODUCT_SERVICE_APP_PORT")
	if port == "" {
		port = "3010"
	}

	log.Printf("Product service is running on port %s", port)
	log.Fatal(app.Listen(":" + port))
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
