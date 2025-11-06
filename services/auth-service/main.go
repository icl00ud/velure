package main

import (
	"log"
	"net/http"
	"os"

	"velure-auth-service/internal/config"
	"velure-auth-service/internal/database"
	"velure-auth-service/internal/handlers"
	"velure-auth-service/internal/middleware"
	"velure-auth-service/internal/repositories"
	"velure-auth-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	log.Println("Loading configuration...")
	cfg := config.Load()
	log.Printf("Configuration loaded successfully")
	log.Printf("Database config - Host: %s, Port: %d, Database: %s, User: %s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.Database, cfg.Database.Username)

	// Initialize database
	log.Println("Connecting to database...")
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connection established successfully")

	// Auto migrate
	log.Println("Running database migrations...")
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migrations completed successfully")

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	passwordResetRepo := repositories.NewPasswordResetRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, sessionRepo, passwordResetRepo, cfg)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Set gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := gin.Default()

	// Middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.PrometheusMiddleware())

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Routes
	setupRoutes(router, authHandler)

	// Start server
	port := os.Getenv("AUTH_SERVICE_APP_PORT")
	if port == "" {
		port = "3020"
	}

	log.Printf("Authentication service is running on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(router *gin.Engine, authHandler *handlers.AuthHandler) {
	auth := router.Group("/authentication")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/validateToken", authHandler.ValidateToken)
		auth.GET("/users", authHandler.GetUsers)
		auth.GET("/user/id/:id", authHandler.GetUserByID)
		auth.GET("/user/email/:email", authHandler.GetUserByEmail)
		auth.DELETE("/logout/:refreshToken", authHandler.Logout)
	}
}
