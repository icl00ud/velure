package handler

import (
	"context"
	"encoding/json"
	"time"

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
			ID    string `json:"id"`
			Total int    `json:"total"`
		}
		if err := json.Unmarshal(evt.Payload, &p); err != nil {
			return err
		}
		return oc.svc.Process(p.ID, p.Total)
	}

	for i := 0; i < oc.workers; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					oc.logger.Error("worker panic", zap.Int("id", id), zap.Any("panic", r))
				}
			}()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					if err := oc.consumer.Consume(ctx, handler); err != nil {
						if ctx.Err() != nil {
							return
						}
						oc.logger.Error("worker error", zap.Int("id", id), zap.Error(err))
						select {
						case <-ctx.Done():
							return
						case <-time.After(time.Second):
						}
					}
				}
			}
		}(i)
	}
	oc.logger.Info("order consumer started", zap.Int("workers", oc.workers))
	<-ctx.Done()
	return nil
}
