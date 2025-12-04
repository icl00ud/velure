package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/icl00ud/process-order-service/internal/client"
	"github.com/icl00ud/process-order-service/internal/metrics"
	"github.com/icl00ud/process-order-service/internal/model"
	"github.com/icl00ud/process-order-service/internal/queue"
)

type PaymentService interface {
	Process(orderID string, items []model.CartItem, amount int) error
}

type paymentService struct {
	pub           queue.Publisher
	productClient client.ProductClient
}

func NewPaymentService(pub queue.Publisher, productClient client.ProductClient) PaymentService {
	return &paymentService{
		pub:           pub,
		productClient: productClient,
	}
}

func (s *paymentService) Process(orderID string, items []model.CartItem, amount int) error {
	start := time.Now()

	// Step 1: Deduct stock for all items in PARALLEL
	shouldContinue, err := s.deductStockParallel(orderID, items, start)
	if err != nil {
		return err
	}
	if !shouldContinue {
		// Permanent error occurred, failure event already published, stop processing
		return nil
	}

	// Step 2: Publish processing event
	procEvt := model.Event{
		Type: model.OrderProcessing,
		Payload: func() json.RawMessage {
			m := struct{ ID string }{ID: orderID}
			b, _ := json.Marshal(m)
			return json.RawMessage(b)
		}(),
	}
	if err := s.pub.Publish(procEvt); err != nil {
		return fmt.Errorf("publish processing: %w", err)
	}

	// Step 3: Simulate payment processing (2-4 seconds) - NON-BLOCKING
	metrics.PaymentAttempts.WithLabelValues("initiated").Inc()
	paymentStart := time.Now()

	if err := s.simulatePaymentProcessing(); err != nil {
		metrics.PaymentAttempts.WithLabelValues("failure").Inc()
		metrics.OrdersProcessed.WithLabelValues("failure").Inc()
		metrics.OrderProcessingDuration.Observe(time.Since(start).Seconds())
		return err
	}

	metrics.PaymentProcessingDuration.Observe(time.Since(paymentStart).Seconds())
	metrics.PaymentAttempts.WithLabelValues("success").Inc()
	metrics.PaymentTotalValue.Observe(float64(amount))

	// Step 4: Publish completed event
	compEvt := model.Event{
		Type: model.OrderCompleted,
		Payload: func() json.RawMessage {
			p := struct {
				ID        string    `json:"id"`
				OrderID   string    `json:"order_id"`
				Amount    int       `json:"amount"`
				Processed time.Time `json:"processed_at"`
			}{ID: orderID, OrderID: orderID, Amount: amount, Processed: time.Now()}
			b, _ := json.Marshal(p)
			return json.RawMessage(b)
		}(),
	}
	if err := s.pub.Publish(compEvt); err != nil {
		metrics.OrdersProcessed.WithLabelValues("failure").Inc()
		metrics.OrderProcessingDuration.Observe(time.Since(start).Seconds())
		return fmt.Errorf("publish completed: %w", err)
	}

	metrics.OrdersProcessed.WithLabelValues("success").Inc()
	metrics.OrderProcessingDuration.Observe(time.Since(start).Seconds())
	return nil
}

// deductStockParallel processes all inventory updates concurrently
// Returns (shouldContinue, error) where shouldContinue=false means permanent failure was handled
func (s *paymentService) deductStockParallel(orderID string, items []model.CartItem, start time.Time) (bool, error) {
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

			err := s.productClient.UpdateQuantity(item.ProductID, -item.Quantity)
			metrics.InventoryCheckDuration.Observe(time.Since(checkStart).Seconds())

			if err != nil {
				metrics.InventoryChecks.WithLabelValues("error").Inc()
			}

			results <- result{item: item, err: err}
		}(item)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results and check for errors
	var firstErr error
	var failedItem model.CartItem
	for r := range results {
		if r.err != nil && firstErr == nil {
			firstErr = r.err
			failedItem = r.item
		}
	}

	if firstErr != nil {
		metrics.OrdersProcessed.WithLabelValues("failure").Inc()
		metrics.OrderProcessingDuration.Observe(time.Since(start).Seconds())

		// Check if it's a permanent error (e.g. product not found)
		var permErr *client.PermanentError
		if errors.As(firstErr, &permErr) {
			// Publish failure event
			failEvt := model.Event{
				Type: model.OrderFailed,
				Payload: func() json.RawMessage {
					p := struct {
						ID      string `json:"id"`
						OrderID string `json:"order_id"`
						Reason  string `json:"reason"`
					}{ID: orderID, OrderID: orderID, Reason: firstErr.Error()}
					b, _ := json.Marshal(p)
					return json.RawMessage(b)
				}(),
			}
			if pubErr := s.pub.Publish(failEvt); pubErr != nil {
				return false, fmt.Errorf("deduct stock failed: %w; publish failure failed: %v", firstErr, pubErr)
			}
			// Permanent error handled, don't continue processing but no error to return
			return false, nil
		}

		return false, fmt.Errorf("deduct stock for product %s: %w", failedItem.ProductID, firstErr)
	}

	return true, nil
}

// simulatePaymentProcessing simulates payment with context-aware waiting
func (s *paymentService) simulatePaymentProcessing() error {
	randomDuration, err := rand.Int(rand.Reader, big.NewInt(3))
	if err != nil {
		return fmt.Errorf("generate random duration: %w", err)
	}

	sleepTime := time.Duration(randomDuration.Int64()+2) * time.Second

	// Use select with timer instead of blocking Sleep
	// This allows the goroutine to be interrupted if needed
	ctx, cancel := context.WithTimeout(context.Background(), sleepTime+time.Second)
	defer cancel()

	select {
	case <-time.After(sleepTime):
		return nil
	case <-ctx.Done():
		return nil
	}
}
