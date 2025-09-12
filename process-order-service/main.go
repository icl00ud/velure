package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/icl00ud/process-order-service/internal/config"
	"github.com/icl00ud/process-order-service/internal/handler"
	"github.com/icl00ud/process-order-service/internal/queue"
	"github.com/icl00ud/process-order-service/internal/service"
	"github.com/joho/godotenv"
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

	var consumer queue.Consumer
	var publisher queue.Publisher

	for i := 0; i < 5; i++ {
		consumer, err = queue.NewRabbitMQConsumer(cfg.RabbitURL, cfg.OrderQueue, logger)
		if err != nil {
			logger.Warn("consumer init failed, retrying", zap.Error(err), zap.Int("attempt", i+1))
			time.Sleep(time.Duration(i+1) * 2 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		logger.Fatal("consumer init failed after retries", zap.Error(err))
	}
	defer consumer.Close()

	for i := 0; i < 5; i++ {
		publisher, err = queue.NewRabbitPublisher(cfg.RabbitURL, cfg.OrderExchange, logger)
		if err != nil {
			logger.Warn("publisher init failed, retrying", zap.Error(err), zap.Int("attempt", i+1))
			time.Sleep(time.Duration(i+1) * 2 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		logger.Fatal("publisher init failed after retries", zap.Error(err))
	}
	defer publisher.Close()

	paySvc := service.NewPaymentService(publisher)
	oc := handler.NewOrderConsumer(consumer, paySvc, cfg.Workers, logger)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return oc.Start(ctx)
	})

	if err := g.Wait(); err != nil && err != context.Canceled {
		logger.Fatal("error during execution", zap.Error(err))
	}
	logger.Info("shutdown complete")
}
