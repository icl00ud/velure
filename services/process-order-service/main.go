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

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("config error", zap.Error(err))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var rabbitConn *queue.RabbitMQConnection
	for i := 0; i < 5; i++ {
		rabbitConn, err = queue.NewRabbitMQConnection(cfg.RabbitURL, logger)
		if err != nil {
			logger.Warn("rabbitmq connection failed, retrying", zap.Error(err), zap.Int("attempt", i+1))
			time.Sleep(time.Duration(i+1) * 2 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		logger.Fatal("rabbitmq connection failed after retries", zap.Error(err))
	}
	defer rabbitConn.Close()

	consumer, err := rabbitConn.NewConsumer(cfg.OrderQueue)
	if err != nil {
		logger.Fatal("consumer init failed", zap.Error(err))
	}
	defer consumer.Close()

	publisher, err := rabbitConn.NewPublisher(cfg.OrderExchange)
	if err != nil {
		logger.Fatal("publisher init failed", zap.Error(err))
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
		port := os.Getenv("PROCESS_ORDER_SERVICE_APP_PORT")
		if port == "" {
			port = "3040"
		}
		mux := http.NewServeMux()
		// Prometheus metrics endpoint
		mux.Handle("/metrics", promhttp.Handler())
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		srv := &http.Server{Addr: ":" + port, Handler: mux}
		go func() {
			<-ctx.Done()
			_ = srv.Shutdown(context.Background())
		}()
		return srv.ListenAndServe()
	})
	g.Go(func() error {
		return oc.Start(ctx)
	})

	if err := g.Wait(); err != nil && err != context.Canceled {
		logger.Fatal("error during execution", zap.Error(err))
	}
	logger.Info("shutdown complete")
}

