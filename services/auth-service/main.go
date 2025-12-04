package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"velure-auth-service/internal/config"
	"velure-auth-service/internal/database"
	"velure-auth-service/internal/handler"
	"velure-auth-service/internal/middleware"
	"velure-auth-service/internal/repository"
	"velure-auth-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/icl00ud/velure-shared/logger"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

func main() {
	log := logger.Init(logger.Config{
		ServiceName: "auth-service",
		Level:       os.Getenv("LOG_LEVEL"),
		UseColor:    os.Getenv("LOG_COLOR") != "false",
	})

	if err := run(log); err != nil {
		log.Fatal("Failed to start server", logger.Err(err))
	}
}

func run(log *logger.Logger) error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Info("No .env file found, using environment variables")
	}

	// Load configuration
	log.Info("Loading configuration")
	cfg := config.Load()
	log.Info("Configuration loaded",
		logger.String("db_host", cfg.Database.Host),
		logger.Int("db_port", cfg.Database.Port),
		logger.String("db_name", cfg.Database.Database))

	// Initialize database
	log.Info("Connecting to database")
	db, err := database.Connect(cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Info("Database connected")

	// Auto migrate
	log.Info("Running migrations")
	if err := database.Migrate(db); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Info("Migrations completed")

	// Initialize Redis
	log.Info("Connecting to Redis", logger.String("addr", cfg.Redis.Addr))
	redisClient, err := connectRedis(cfg.Redis)
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}
	defer redisClient.Close()
	log.Info("Redis connected")

	// Initialize repositories
	log.Info("Initializing repositories")
	userRepo := repositories.NewUserRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	passwordResetRepo := repositories.NewPasswordResetRepository(db)

	// Initialize services
	log.Info("Initializing services")
	authService := services.NewAuthService(userRepo, sessionRepo, passwordResetRepo, cfg, redisClient)
	authService.SyncActiveSessionsMetric(context.Background())
	authService.SyncTotalUsersMetric(context.Background())

	// Start background metrics sync job (updates gauges every 30s)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			authService.SyncActiveSessionsMetric(context.Background())
			authService.SyncTotalUsersMetric(context.Background())
		}
	}()
	log.Info("Background metrics sync started", logger.String("interval", "30s"))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Set up router with all middleware and routes
	log.Info("Setting up HTTP router")
	router := setupRouter(cfg, authHandler)

	// Start server
	port := os.Getenv("AUTH_SERVICE_APP_PORT")
	if port == "" {
		port = "3020"
	}

	log.Info("Starting HTTP server", logger.String("port", port))
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

	// Initialize router without default middleware
	router := gin.New()
	router.Use(gin.Recovery())

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
