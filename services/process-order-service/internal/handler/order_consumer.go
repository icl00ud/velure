package handler

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/icl00ud/process-order-service/internal/model"
	"github.com/icl00ud/process-order-service/internal/queue"
	"github.com/icl00ud/process-order-service/internal/service"
)

type OrderConsumer struct {
	consumer queue.Consumer
	svc      service.PaymentService
	workers  int
	logger   *zap.Logger
}

func NewOrderConsumer(c queue.Consumer, svc service.PaymentService, workers int, logger *zap.Logger) *OrderConsumer {
	return &OrderConsumer{
		consumer: c,
		svc:      svc,
		workers:  workers,
		logger:   logger,
	}
}

func (oc *OrderConsumer) Start(ctx context.Context) error {
	handler := func(evt model.Event) error {
		if evt.Type != model.OrderCreated {
			return nil
		}
		var p struct {
			ID    string            `json:"id"`
			Items []model.CartItem  `json:"items"`
			Total float64           `json:"total"`
		}
		if err := json.Unmarshal(evt.Payload, &p); err != nil {
			return err
		}
		return oc.svc.Process(p.ID, p.Items, int(p.Total))
	}

	oc.logger.Info("order consumer started", zap.Int("workers", oc.workers))

	for i := 0; i < oc.workers; i++ {
		go func(workerID int) {
			defer func() {
				if r := recover(); r != nil {
					oc.logger.Error("worker panic", zap.Int("worker_id", workerID), zap.Any("panic", r))
				}
			}()

			if err := oc.consumer.Consume(ctx, handler); err != nil && ctx.Err() == nil {
				oc.logger.Error("worker stopped", zap.Int("worker_id", workerID), zap.Error(err))
			}
		}(i)
	}

	<-ctx.Done()
	return nil
}
