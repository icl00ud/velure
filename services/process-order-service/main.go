package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/icl00ud/velure/services/process-order-service/internal/client"
	"github.com/icl00ud/velure/services/process-order-service/internal/config"
	"github.com/icl00ud/velure/services/process-order-service/internal/handler"
	"github.com/icl00ud/velure/services/process-order-service/internal/idempotency"
	"github.com/icl00ud/velure/services/process-order-service/internal/payment"
	"github.com/icl00ud/velure/services/process-order-service/internal/queue"
	"github.com/icl00ud/velure/services/process-order-service/internal/service"
	"github.com/icl00ud/velure/services/process-order-service/internal/telemetry"
	"github.com/icl00ud/velure/shared/logger"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

func main() {
	_ = godotenv.Load()
	log := logger.Init(logger.Config{
		ServiceName: "process-order-service",
		Level:       os.Getenv("LOG_LEVEL"),
		UseColor:    os.Getenv("LOG_COLOR") != "false",
	})
	defer log.Sync()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, log); err != nil {
		log.Fatal("error during execution", logger.Err(err))
	}
}

func run(ctx context.Context, log *logger.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	otelShutdown, err := telemetry.Init(ctx, "process-order-service")
	if err != nil {
		log.Warn("telemetry init failed, continuing without tracing", logger.Err(err))
	}
	defer func() {
		shutdownCtx, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()
		_ = otelShutdown(shutdownCtx)
	}()

	var rabbitConn *queue.RabbitMQConnection
	for i := 0; i < 5; i++ {
		rabbitConn, err = queue.NewRabbitMQConnection(cfg.RabbitURL, log)
		if err != nil {
			log.Warn("rabbitmq connection failed, retrying", logger.Err(err), logger.Int("attempt", i+1))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(i+1) * 2 * time.Second):
			}
			continue
		}
		break
	}
	if err != nil {
		return fmt.Errorf("rabbitmq connection failed after retries: %w", err)
	}
	defer rabbitConn.Close()

	consumer, err := rabbitConn.NewConsumer(cfg.OrderQueue)
	if err != nil {
		return fmt.Errorf("consumer init failed: %w", err)
	}
	defer consumer.Close()

	publisher, err := rabbitConn.NewPublisher(cfg.OrderExchange)
	if err != nil {
		return fmt.Errorf("publisher init failed: %w", err)
	}
	defer publisher.Close()

	// Initialize product client
	productServiceURL := os.Getenv("PRODUCT_SERVICE_URL")
	if productServiceURL == "" {
		productServiceURL = "http://product-service:3010"
	}
	productClient := client.NewProductClient(productServiceURL)

	// Stripe test mode when STRIPE_API_KEY is set; otherwise a latency-only
	// simulated processor so local development needs no Stripe account.
	var processor payment.Processor
	if key := os.Getenv("STRIPE_API_KEY"); key != "" {
		log.Info("payment processor: stripe (test mode)")
		processor = payment.NewStripeProcessor(key)
	} else {
		log.Info("payment processor: simulated")
		processor = payment.NewSimulatedProcessor(3 * time.Second)
	}

	paySvc := service.NewPaymentService(publisher, productClient, processor)

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Warn("redis ping failed; idempotency will fail-open", logger.Err(err))
	}
	checker := idempotency.NewChecker(rdb, 24*time.Hour)

	oc := handler.NewOrderConsumer(consumer, paySvc, checker, cfg.Workers, log)

	g, ctx := errgroup.WithContext(ctx)
	// health server
	g.Go(func() error {
		mux := http.NewServeMux()
		// Prometheus metrics endpoint
		mux.Handle("/metrics", promhttp.Handler())
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		srv := &http.Server{Addr: ":" + cfg.Port, Handler: mux}
		go func() {
			<-ctx.Done()
			_ = srv.Shutdown(context.Background())
		}()
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	g.Go(func() error {
		return oc.Start(ctx)
	})

	if err := g.Wait(); err != nil && err != context.Canceled {
		return err
	}
	log.Info("Shutdown complete")
	return nil
}
