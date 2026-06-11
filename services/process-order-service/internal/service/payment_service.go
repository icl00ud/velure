package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/icl00ud/velure/services/process-order-service/internal/client"
	"github.com/icl00ud/velure/services/process-order-service/internal/metrics"
	"github.com/icl00ud/velure/services/process-order-service/internal/model"
	"github.com/icl00ud/velure/services/process-order-service/internal/payment"
	"github.com/icl00ud/velure/services/process-order-service/internal/queue"
	"github.com/icl00ud/velure/shared/logger"
)

type PaymentService interface {
	Process(ctx context.Context, orderID string, items []model.CartItem, amount int) error
}

type paymentService struct {
	pub           queue.Publisher
	productClient client.ProductClient
	processor     payment.Processor
}

func NewPaymentService(pub queue.Publisher, productClient client.ProductClient, processor payment.Processor) PaymentService {
	return &paymentService{
		pub:           pub,
		productClient: productClient,
		processor:     processor,
	}
}

// Process deducts stock, charges the payment and publishes status events.
// Invariant: when Process fails after deducting stock (error return or
// permanent failure), every successful deduction is compensated so a retry —
// which re-runs the whole flow — never double-deducts inventory.
func (s *paymentService) Process(ctx context.Context, orderID string, items []model.CartItem, amount int) error {
	ctx, span := otel.Tracer("process-order").Start(ctx, "order.process",
		trace.WithAttributes(attribute.String("velure.order_id", orderID)))
	defer span.End()

	start := time.Now()

	// Step 1: Deduct stock for all items in PARALLEL
	deducted, firstErr, failedItem := s.deductStockParallel(ctx, items)
	if firstErr != nil {
		s.compensateStock(ctx, orderID, deducted)
		metrics.OrdersProcessed.WithLabelValues("failure").Inc()
		metrics.OrderProcessingDuration.Observe(time.Since(start).Seconds())

		// Permanent errors (e.g. product not found) are not retryable:
		// publish the failure event and ack the message.
		var permErr *client.PermanentError
		if errors.As(firstErr, &permErr) {
			failEvt := model.Event{
				Type: model.OrderFailed,
				Payload: mustJSON(struct {
					ID      string `json:"id"`
					OrderID string `json:"order_id"`
					Reason  string `json:"reason"`
				}{ID: orderID, OrderID: orderID, Reason: firstErr.Error()}),
			}
			if pubErr := s.pub.Publish(ctx, failEvt); pubErr != nil {
				return fmt.Errorf("deduct stock failed: %w; publish failure failed: %v", firstErr, pubErr)
			}
			return nil
		}

		return fmt.Errorf("deduct stock for product %s: %w", failedItem.ProductID, firstErr)
	}

	// Step 2: Publish processing event
	procEvt := model.Event{
		Type: model.OrderProcessing,
		Payload: mustJSON(struct {
			ID string
		}{ID: orderID}),
	}
	if err := s.pub.Publish(ctx, procEvt); err != nil {
		s.compensateStock(ctx, orderID, items)
		return fmt.Errorf("publish processing: %w", err)
	}

	// Step 3: Charge the payment (Stripe test mode, or the simulated
	// processor when no API key is configured). The order ID doubles as the
	// idempotency key, so retries cannot double-charge.
	metrics.PaymentAttempts.WithLabelValues("initiated").Inc()
	paymentStart := time.Now()

	if err := s.processor.Charge(ctx, orderID, int64(amount)); err != nil {
		s.compensateStock(ctx, orderID, items)
		metrics.PaymentAttempts.WithLabelValues("failure").Inc()
		metrics.OrdersProcessed.WithLabelValues("failure").Inc()
		metrics.OrderProcessingDuration.Observe(time.Since(start).Seconds())

		// Declined payments are final: fail the order and ack the message.
		var payErr *payment.PermanentError
		if errors.As(err, &payErr) {
			failEvt := model.Event{
				Type: model.OrderFailed,
				Payload: mustJSON(struct {
					ID      string `json:"id"`
					OrderID string `json:"order_id"`
					Reason  string `json:"reason"`
				}{ID: orderID, OrderID: orderID, Reason: payErr.Reason}),
			}
			if pubErr := s.pub.Publish(ctx, failEvt); pubErr != nil {
				return fmt.Errorf("payment failed: %w; publish failure failed: %v", err, pubErr)
			}
			return nil
		}
		return fmt.Errorf("charge payment: %w", err)
	}

	metrics.PaymentProcessingDuration.Observe(time.Since(paymentStart).Seconds())
	metrics.PaymentAttempts.WithLabelValues("success").Inc()
	metrics.PaymentTotalValue.Observe(float64(amount))

	// Step 4: Publish completed event
	compEvt := model.Event{
		Type: model.OrderCompleted,
		Payload: mustJSON(struct {
			ID        string    `json:"id"`
			OrderID   string    `json:"order_id"`
			Amount    int       `json:"amount"`
			Processed time.Time `json:"processed_at"`
		}{ID: orderID, OrderID: orderID, Amount: amount, Processed: time.Now()}),
	}
	if err := s.pub.Publish(ctx, compEvt); err != nil {
		// Error return leads to a retry that re-runs the whole flow, so the
		// stock from this attempt has to be handed back first.
		s.compensateStock(ctx, orderID, items)
		metrics.OrdersProcessed.WithLabelValues("failure").Inc()
		metrics.OrderProcessingDuration.Observe(time.Since(start).Seconds())
		return fmt.Errorf("publish completed: %w", err)
	}

	metrics.OrdersProcessed.WithLabelValues("success").Inc()
	metrics.OrderProcessingDuration.Observe(time.Since(start).Seconds())
	return nil
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return json.RawMessage(b)
}

// deductStockParallel processes all inventory updates concurrently.
// Returns the items whose deduction succeeded, the first error encountered
// and the item it belongs to.
func (s *paymentService) deductStockParallel(ctx context.Context, items []model.CartItem) ([]model.CartItem, error, model.CartItem) {
	type result struct {
		item model.CartItem
		err  error
	}

	results := make(chan result, len(items))
	var wg sync.WaitGroup

	for _, item := range items {
		wg.Add(1)
		go func(item model.CartItem) {
			defer wg.Done()

			metrics.InventoryChecks.WithLabelValues("available").Inc()
			checkStart := time.Now()

			err := s.productClient.UpdateQuantity(ctx, item.ProductID, -item.Quantity)
			metrics.InventoryCheckDuration.Observe(time.Since(checkStart).Seconds())

			if err != nil {
				metrics.InventoryChecks.WithLabelValues("error").Inc()
			}

			results <- result{item: item, err: err}
		}(item)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	deducted := make([]model.CartItem, 0, len(items))
	var firstErr error
	var failedItem model.CartItem
	for r := range results {
		if r.err != nil {
			if firstErr == nil {
				firstErr = r.err
				failedItem = r.item
			}
			continue
		}
		deducted = append(deducted, r.item)
	}

	return deducted, firstErr, failedItem
}

// compensateStock re-adds quantities that were successfully deducted in a
// failed processing attempt. Compensation failures are logged but not
// propagated — there is no further recovery at this layer.
func (s *paymentService) compensateStock(ctx context.Context, orderID string, items []model.CartItem) {
	for _, item := range items {
		if err := s.productClient.UpdateQuantity(ctx, item.ProductID, item.Quantity); err != nil {
			metrics.InventoryChecks.WithLabelValues("error").Inc()
			logger.Error("stock compensation failed — manual reconciliation needed",
				logger.String("order_id", orderID),
				logger.String("product_id", item.ProductID),
				logger.Int("quantity", item.Quantity),
				logger.Err(err))
		}
	}
}
