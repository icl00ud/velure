package handler

import (
	"context"
	"encoding/json"

	"github.com/icl00ud/velure/shared/logger"

	"github.com/icl00ud/velure/services/process-order-service/internal/metrics"
	"github.com/icl00ud/velure/services/process-order-service/internal/model"
	"github.com/icl00ud/velure/services/process-order-service/internal/queue"
	"github.com/icl00ud/velure/services/process-order-service/internal/service"
)

// IdempotencyChecker gates duplicate event processing via Redis SET NX EX.
type IdempotencyChecker interface {
	FirstSeen(ctx context.Context, eventID string) (bool, error)
	Forget(ctx context.Context, eventID string) error
}

type OrderConsumer struct {
	consumer queue.Consumer
	svc      service.PaymentService
	idem     IdempotencyChecker
	workers  int
	logger   *logger.Logger
}

func NewOrderConsumer(c queue.Consumer, svc service.PaymentService, idem IdempotencyChecker, workers int, log *logger.Logger) *OrderConsumer {
	return &OrderConsumer{
		consumer: c,
		svc:      svc,
		idem:     idem,
		workers:  workers,
		logger:   log,
	}
}

func (oc *OrderConsumer) Start(ctx context.Context) error {
	handler := func(ctx context.Context, eventID string, evt model.Event) error {
		metrics.MessagesConsumed.Inc()

		// Idempotency gate (fail-open on Redis errors).
		if oc.idem != nil && eventID != "" {
			firstSeen, err := oc.idem.FirstSeen(ctx, eventID)
			if err != nil {
				oc.logger.Error("idempotency check failed, processing anyway", logger.Err(err))
				metrics.IdempotencyCheckFailed.Inc()
			} else if !firstSeen {
				metrics.DuplicatesSkipped.Inc()
				metrics.MessagesAcknowledged.WithLabelValues("ack").Inc()
				return nil // ack and skip
			}
		}

		if evt.Type != model.OrderCreated {
			metrics.MessagesAcknowledged.WithLabelValues("ack").Inc()
			return nil
		}

		var p struct {
			ID    string           `json:"id"`
			Items []model.CartItem `json:"items"`
			Total float64          `json:"total"`
		}
		if err := json.Unmarshal(evt.Payload, &p); err != nil {
			metrics.MessageProcessingErrors.Inc()
			metrics.MessagesAcknowledged.WithLabelValues("nack").Inc()
			if oc.idem != nil && eventID != "" {
				_ = oc.idem.Forget(ctx, eventID)
			}
			return err
		}

		if err := oc.svc.Process(ctx, p.ID, p.Items, int(p.Total)); err != nil {
			metrics.MessageProcessingErrors.Inc()
			metrics.MessagesAcknowledged.WithLabelValues("nack").Inc()
			if oc.idem != nil && eventID != "" {
				_ = oc.idem.Forget(ctx, eventID)
			}
			return err
		}

		metrics.MessagesAcknowledged.WithLabelValues("ack").Inc()
		return nil
	}

	oc.logger.Info("order consumer started", logger.Int("workers", oc.workers))
	metrics.ActiveWorkers.Set(float64(oc.workers))

	for i := 0; i < oc.workers; i++ {
		go func(workerID int) {
			defer func() {
				if r := recover(); r != nil {
					oc.logger.Error("worker panic", logger.Int("worker_id", workerID), logger.Any("panic", r))
					metrics.Errors.WithLabelValues("internal").Inc()
				}
				metrics.ActiveWorkers.Dec()
			}()

			if err := oc.consumer.Consume(ctx, handler); err != nil && ctx.Err() == nil {
				oc.logger.Error("worker stopped", logger.Int("worker_id", workerID), logger.Err(err))
				metrics.Errors.WithLabelValues("rabbitmq").Inc()
			}
		}(i)
	}

	<-ctx.Done()
	return nil
}
