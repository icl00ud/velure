package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/icl00ud/publish-order-service/internal/config"
	"github.com/icl00ud/publish-order-service/internal/consumer"
	"github.com/icl00ud/publish-order-service/internal/database"
	"github.com/icl00ud/publish-order-service/internal/handler"
	"github.com/icl00ud/publish-order-service/internal/middleware"
	"github.com/icl00ud/publish-order-service/internal/publisher"
	"github.com/icl00ud/publish-order-service/internal/repository"
	"github.com/icl00ud/publish-order-service/internal/service"
	"github.com/icl00ud/velure-shared/logger"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type dbProvider interface {
	DB() *sql.DB
}

type server interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

type appDeps struct {
	loadConfig    func() (config.Config, error)
	newRepo       func(string) (repository.OrderRepository, error)
	runMigrations func(*sql.DB, string) error
	newPublisher  func(string, string, *logger.Logger) (publisher.Publisher, error)
	newConsumer   func(string, string, string, consumer.EventHandler, int, *logger.Logger) (consumer.Consumer, error)
	newLogger     func() *logger.Logger
	newHTTPServer func(config.Config, http.Handler) server
}

var depsFactory = defaultDeps

func defaultDeps() appDeps {
	return appDeps{
		loadConfig:    config.Load,
		newRepo:       repository.NewOrderRepository,
		runMigrations: database.RunMigrations,
		newPublisher:  publisher.NewRabbitMQPublisher,
		newConsumer:   consumer.NewRabbitMQConsumer,
		newLogger: func() *logger.Logger {
			return logger.Init(logger.Config{
				ServiceName: "publish-order-service",
				Level:       os.Getenv("LOG_LEVEL"),
				UseColor:    os.Getenv("LOG_COLOR") != "false",
			})
		},
		newHTTPServer: func(cfg config.Config, handler http.Handler) server {
			return &http.Server{
				Addr:         ":" + cfg.Port,
				Handler:      handler,
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  120 * time.Second,
			}
		},
	}
}

func main() {
	if err := run(context.Background(), depsFactory()); err != nil {
		logger.Fatal("error during execution", logger.Err(err))
	}
}

func run(parentCtx context.Context, deps appDeps) error {
	godotenv.Load()

	log := deps.newLogger()
	defer log.Sync()

	log.Info("Starting publish-order-service")

	log.Info("Loading configuration")
	cfg, err := deps.loadConfig()
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}
	log.Info("Configuration loaded",
		logger.String("port", cfg.Port),
		logger.String("exchange", cfg.Exchange),
		logger.String("queue", cfg.Queue),
		logger.Int("workers", cfg.Workers))

	ctx, cancel := signal.NotifyContext(parentCtx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Info("Connecting to PostgreSQL")
	repo, err := deps.newRepo(cfg.PostgresURL)
	if err != nil {
		return fmt.Errorf("repository init failed: %w", err)
	}
	log.Info("PostgreSQL connected")

	if dbRepo, ok := repo.(dbProvider); ok {
		log.Info("Running migrations")
		if err := deps.runMigrations(dbRepo.DB(), "./internal/migrations"); err != nil {
			return fmt.Errorf("migration error: %w", err)
		}
		log.Info("Migrations completed")
	} else {
		log.Info("Skipping migrations")
	}

	log.Info("Connecting to RabbitMQ", logger.String("exchange", cfg.Exchange))
	pub, err := deps.newPublisher(cfg.RabbitURL, cfg.Exchange, log)
	if err != nil {
		return fmt.Errorf("publisher init failed: %w", err)
	}
	defer pub.Close()
	log.Info("RabbitMQ publisher initialized")

	log.Info("Initializing services")
	svc := service.NewOrderService(repo, service.NewPricingCalculator())
	oh := handler.NewOrderHandler(svc, pub)

	sseHandler := handler.NewSSEHandler(svc)
	eventHandler := handler.NewEventHandler(svc, log)
	eventHandler.SetSSEHandler(sseHandler)

	log.Info("Initializing RabbitMQ consumer", logger.String("queue", cfg.Queue), logger.Int("workers", cfg.Workers))
	cons, err := deps.newConsumer(
		cfg.RabbitURL,
		cfg.Exchange,
		cfg.Queue,
		eventHandler.HandleEvent,
		cfg.Workers,
		log,
	)
	if err != nil {
		return fmt.Errorf("consumer init failed: %w", err)
	}
	defer cons.Close()

	log.Info("Setting up middleware")
	authMiddleware := middleware.Auth(cfg.JWTSecret)
	sseAuthMiddleware := middleware.SSEAuth(cfg.JWTSecret)

	mux := http.NewServeMux()
	// Support both root paths (local dev with Caddy rewrite) and /api/order (Kubernetes ALB)
	mux.Handle("/create-order", middleware.CORS(middleware.Logging(middleware.Timeout(5*time.Second)(authMiddleware(http.HandlerFunc(oh.CreateOrder))))))
	mux.Handle("/update-order-status", middleware.CORS(middleware.Logging(middleware.Timeout(5*time.Second)(http.HandlerFunc(oh.UpdateStatus)))))
	mux.Handle("/orders", middleware.CORS(middleware.Logging(middleware.Timeout(3*time.Second)(http.HandlerFunc(oh.GetOrdersByPage)))))
	mux.Handle("/user/orders", middleware.CORS(middleware.Logging(middleware.Timeout(3*time.Second)(authMiddleware(http.HandlerFunc(oh.GetUserOrders))))))
	mux.Handle("/user/order", middleware.CORS(middleware.Logging(middleware.Timeout(3*time.Second)(authMiddleware(http.HandlerFunc(oh.GetUserOrderByID))))))
	mux.Handle("/user/order/status", middleware.CORS(middleware.Logging(sseAuthMiddleware(http.HandlerFunc(sseHandler.StreamOrderStatus)))))

	// Kubernetes ALB routes (no path rewriting)
	mux.Handle("/api/order/create-order", middleware.CORS(middleware.Logging(middleware.Timeout(5*time.Second)(authMiddleware(http.HandlerFunc(oh.CreateOrder))))))
	mux.Handle("/api/order/update-order-status", middleware.CORS(middleware.Logging(middleware.Timeout(5*time.Second)(http.HandlerFunc(oh.UpdateStatus)))))
	mux.Handle("/api/order/orders", middleware.CORS(middleware.Logging(middleware.Timeout(3*time.Second)(http.HandlerFunc(oh.GetOrdersByPage)))))
	mux.Handle("/api/order/user/orders", middleware.CORS(middleware.Logging(middleware.Timeout(3*time.Second)(authMiddleware(http.HandlerFunc(oh.GetUserOrders))))))
	mux.Handle("/api/order/user/order", middleware.CORS(middleware.Logging(middleware.Timeout(3*time.Second)(authMiddleware(http.HandlerFunc(oh.GetUserOrderByID))))))
	mux.Handle("/api/order/user/order/status", middleware.CORS(middleware.Logging(sseAuthMiddleware(http.HandlerFunc(sseHandler.StreamOrderStatus)))))

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

	log.Info("Starting HTTP server", logger.String("port", cfg.Port))
	srv := deps.newHTTPServer(cfg, mux)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server error", logger.Err(err))
			return err
		}
		return nil
	})

	g.Go(func() error {
		log.Info("Consumer starting", logger.String("queue", cfg.Queue), logger.Int("workers", cfg.Workers))
		return cons.Start(ctx)
	})

	g.Go(func() error {
		<-ctx.Done()
		log.Info("Shutting down server")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		return srv.Shutdown(shutdownCtx)
	})

	if err := g.Wait(); err != nil && err != context.Canceled {
		return err
	}

	log.Info("Shutdown complete")
	return nil
}
