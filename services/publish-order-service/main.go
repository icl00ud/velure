package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/icl00ud/publish-order-service/internal/config"
	"github.com/icl00ud/publish-order-service/internal/consumer"
	"github.com/icl00ud/publish-order-service/internal/database"
	"github.com/icl00ud/publish-order-service/internal/handler"
	"github.com/icl00ud/publish-order-service/internal/middleware"
	"github.com/icl00ud/publish-order-service/internal/publisher"
	"github.com/icl00ud/publish-order-service/internal/repository"
	"github.com/icl00ud/publish-order-service/internal/service"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	godotenv.Load()
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	logger.Info("Starting publish-order-service initialization...")

	logger.Info("Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("config error", zap.Error(err))
	}
	logger.Info("Configuration loaded successfully",
		zap.String("port", cfg.Port),
		zap.String("exchange", cfg.Exchange),
		zap.String("queue", cfg.Queue),
		zap.Int("workers", cfg.Workers))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger.Info("Connecting to PostgreSQL database...", zap.String("postgres_url_masked", "postgresql://***"))
	repo, err := repository.NewOrderRepository(cfg.PostgresURL)
	if err != nil {
		logger.Fatal("repository init failed", zap.Error(err))
	}
	logger.Info("PostgreSQL database connection established successfully")

	logger.Info("Running database migrations...")
	if err := database.RunMigrations(repo.(*repository.PostgresOrderRepository).DB(), "./migrations"); err != nil {
		logger.Fatal("migration error", zap.Error(err))
	}
	logger.Info("Database migrations completed successfully")

	logger.Info("Connecting to RabbitMQ...", zap.String("exchange", cfg.Exchange))
	pub, err := publisher.NewRabbitMQPublisher(cfg.RabbitURL, cfg.Exchange, logger)
	if err != nil {
		logger.Fatal("publisher init failed", zap.Error(err))
	}
	defer pub.Close()
	logger.Info("RabbitMQ publisher initialized successfully")

	logger.Info("Initializing services and handlers...")
	svc := service.NewOrderService(repo, service.NewPricingCalculator())
	oh := handler.NewOrderHandler(svc, pub)

	sseHandler := handler.NewSSEHandler(svc)
	eventHandler := handler.NewEventHandler(svc, logger)
	eventHandler.SetSSEHandler(sseHandler)
	logger.Info("Services and handlers initialized successfully")

	logger.Info("Initializing RabbitMQ consumer...", zap.String("queue", cfg.Queue), zap.Int("workers", cfg.Workers))
	cons, err := consumer.NewRabbitMQConsumer(
		cfg.RabbitURL,
		cfg.Exchange,
		cfg.Queue,
		eventHandler.HandleEvent,
		cfg.Workers,
		logger,
	)
	if err != nil {
		logger.Fatal("consumer init failed", zap.Error(err))
	}
	defer cons.Close()
	logger.Info("RabbitMQ consumer initialized successfully")

	logger.Info("Setting up middleware...")
	authMiddleware := middleware.Auth(cfg.JWTSecret)
	sseAuthMiddleware := middleware.SSEAuth(cfg.JWTSecret)
	logger.Info("Middleware configured successfully")

	mux := http.NewServeMux()
	mux.Handle("/create-order", middleware.CORS(middleware.Logging(middleware.Timeout(5*time.Second)(authMiddleware(http.HandlerFunc(oh.CreateOrder))))))
	mux.Handle("/update-order-status", middleware.CORS(middleware.Logging(middleware.Timeout(5*time.Second)(http.HandlerFunc(oh.UpdateStatus)))))
	mux.Handle("/orders", middleware.CORS(middleware.Logging(middleware.Timeout(3*time.Second)(http.HandlerFunc(oh.GetOrdersByPage)))))
	mux.Handle("/user/orders", middleware.CORS(middleware.Logging(middleware.Timeout(3*time.Second)(authMiddleware(http.HandlerFunc(oh.GetUserOrders))))))
	mux.Handle("/user/order", middleware.CORS(middleware.Logging(middleware.Timeout(3*time.Second)(authMiddleware(http.HandlerFunc(oh.GetUserOrderByID))))))
	mux.Handle("/user/order/status", middleware.CORS(middleware.Logging(sseAuthMiddleware(http.HandlerFunc(sseHandler.StreamOrderStatus)))))
	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	logger.Info("Configuring HTTP server routes...")
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	logger.Info("HTTP server configured, ready to start", zap.String("port", cfg.Port))

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logger.Info("HTTP server starting...", zap.String("addr", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", zap.Error(err))
			return err
		}
		return nil
	})

	g.Go(func() error {
		logger.Info("consumer starting", zap.String("queue", cfg.Queue), zap.Int("workers", cfg.Workers))
		return cons.Start(ctx)
	})

	g.Go(func() error {
		<-ctx.Done()
		logger.Info("shutting down server")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		return srv.Shutdown(shutdownCtx)
	})

	if err := g.Wait(); err != nil && err != context.Canceled {
		logger.Fatal("error during execution", zap.Error(err))
	}

	logger.Info("shutdown complete")
}
