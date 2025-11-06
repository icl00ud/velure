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
	log.Println("Initializing repositories...")
	userRepo := repositories.NewUserRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	passwordResetRepo := repositories.NewPasswordResetRepository(db)
	log.Println("Repositories initialized successfully")

	// Initialize services
	log.Println("Initializing services...")
	authService := services.NewAuthService(userRepo, sessionRepo, passwordResetRepo, cfg)
	log.Println("Services initialized successfully")

	// Initialize handlers
	log.Println("Initializing handlers...")
	authHandler := handlers.NewAuthHandler(authService)
	log.Println("Handlers initialized successfully")

	// Set gin mode
	if cfg.Environment == "production" {
		log.Println("Setting Gin to release mode...")
		gin.SetMode(gin.ReleaseMode)
	} else {
		log.Println("Running in development mode...")
	}

	// Initialize router
	log.Println("Initializing HTTP router...")
	router := gin.Default()

	// Middleware
	log.Println("Configuring middleware...")
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.PrometheusMiddleware())
	log.Println("Middleware configured successfully")

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Routes
	log.Println("Setting up API routes...")
	setupRoutes(router, authHandler)
	log.Println("API routes configured successfully")

	// Start server
	port := os.Getenv("AUTH_SERVICE_APP_PORT")
	if port == "" {
		port = "3020"
	}

	log.Printf("Starting HTTP server on port %s...", port)
	log.Println("Authentication service initialization completed successfully")
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
