package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"

	"github.com/icl00ud/publish-order-service/internal/config"
	"github.com/icl00ud/publish-order-service/internal/handler"
	"github.com/icl00ud/publish-order-service/internal/middleware"
	"github.com/icl00ud/publish-order-service/internal/publisher"
	"github.com/icl00ud/publish-order-service/internal/repository"
	"github.com/icl00ud/publish-order-service/internal/service"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("config error", zap.Error(err))
	}

	repo, _ := repository.NewOrderRepository(cfg.PostgresURL)
	pub, _ := publisher.NewRabbitMQPublisher(cfg.RabbitURL, cfg.Exchange, logger)

	svc := service.NewOrderService(repo, service.NewPricingCalculator())
	oh := handler.NewOrderHandler(svc, pub)

	mux := http.NewServeMux()
	mux.Handle("/create-order", middleware.Logging(http.HandlerFunc(oh.CreateOrder)))
	mux.Handle("/update-order-status", middleware.Logging(http.HandlerFunc(oh.UpdateStatus)))

	srv := &http.Server{
		Addr: ":" + cfg.Port, Handler: mux,
	}

	go func() {
		zap.L().Info("server starting", zap.String("addr", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("server error", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	zap.L().Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	zap.L().Info("server stopped")
}
