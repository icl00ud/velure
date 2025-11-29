package main

import (
	"context"
	"fmt"
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
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func run() error {
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
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Println("Database connection established successfully")

	// Auto migrate
	log.Println("Running database migrations...")
	if err := database.Migrate(db); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Println("Database migrations completed successfully")

	// Initialize Redis
	log.Println("Connecting to Redis...")
	redisClient, err := connectRedis(cfg.Redis)
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}
	defer redisClient.Close()
	log.Printf("Redis connection established successfully at %s", cfg.Redis.Addr)

	// Initialize repositories
	log.Println("Initializing repositories...")
	userRepo := repositories.NewUserRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	passwordResetRepo := repositories.NewPasswordResetRepository(db)
	log.Println("Repositories initialized successfully")

	// Initialize services
	log.Println("Initializing services...")
	authService := services.NewAuthService(userRepo, sessionRepo, passwordResetRepo, cfg, redisClient)
	authService.SyncActiveSessionsMetric(context.Background())
	authService.SyncTotalUsersMetric(context.Background())
	log.Println("Services initialized successfully")

	// Initialize handlers
	log.Println("Initializing handlers...")
	authHandler := handlers.NewAuthHandler(authService)
	log.Println("Handlers initialized successfully")

	// Set up router with all middleware and routes
	log.Println("Initializing HTTP router...")
	router := setupRouter(cfg, authHandler)
	log.Println("HTTP router initialized successfully")

	// Start server
	port := os.Getenv("AUTH_SERVICE_APP_PORT")
	if port == "" {
		port = "3020"
	}

	log.Printf("Starting HTTP server on port %s...", port)
	log.Println("Authentication service initialization completed successfully")
	if os.Getenv("AUTH_SERVICE_SKIP_HTTP") == "true" {
		return nil
	}

	if err := router.Run(":" + port); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func connectRedis(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

func setupRoutes(router *gin.Engine, authHandler *handlers.AuthHandler) {
	// Support both /authentication (local dev with Caddy rewrite) and /api/auth (Kubernetes ALB)
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

	// Kubernetes ALB routes (no path rewriting)
	apiAuth := router.Group("/api/auth")
	{
		apiAuth.POST("/register", authHandler.Register)
		apiAuth.POST("/login", authHandler.Login)
		apiAuth.POST("/validateToken", authHandler.ValidateToken)
		apiAuth.GET("/users", authHandler.GetUsers)
		apiAuth.GET("/user/id/:id", authHandler.GetUserByID)
		apiAuth.GET("/user/email/:email", authHandler.GetUserByEmail)
		apiAuth.DELETE("/logout/:refreshToken", authHandler.Logout)
	}
}

func setupRouter(cfg *config.Config, authHandler *handlers.AuthHandler) *gin.Engine {
	// Set gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := gin.Default()

	// Rate limiter global
	rateLimiter := middleware.NewRateLimiter(100, 200) // 100 req/s, burst 200

	// Middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.PrometheusMiddleware())
	router.Use(rateLimiter.Middleware())

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Routes
	setupRoutes(router, authHandler)

	return router
}
