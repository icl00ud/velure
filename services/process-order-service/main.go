package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/icl00ud/process-order-service/internal/client"
	"github.com/icl00ud/process-order-service/internal/config"
	"github.com/icl00ud/process-order-service/internal/handler"
	"github.com/icl00ud/process-order-service/internal/queue"
	"github.com/icl00ud/process-order-service/internal/service"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	_ = godotenv.Load()
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, logger); err != nil {
		logger.Fatal("error during execution", zap.Error(err))
	}
}

func run(ctx context.Context, logger *zap.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	var rabbitConn *queue.RabbitMQConnection
	for i := 0; i < 5; i++ {
		rabbitConn, err = queue.NewRabbitMQConnection(cfg.RabbitURL, logger)
		if err != nil {
			logger.Warn("rabbitmq connection failed, retrying", zap.Error(err), zap.Int("attempt", i+1))
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

	paySvc := service.NewPaymentService(publisher, productClient)
	oc := handler.NewOrderConsumer(consumer, paySvc, cfg.Workers, logger)

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
	logger.Info("shutdown complete")
	return nil
}
