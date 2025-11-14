package service

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
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

	// Step 1: Deduct stock for all items BEFORE processing payment
	for _, item := range items {
		metrics.InventoryChecks.WithLabelValues("available").Inc()
		checkStart := time.Now()

		if err := s.productClient.UpdateQuantity(item.ProductID, -item.Quantity); err != nil {
			metrics.InventoryChecks.WithLabelValues("error").Inc()
			metrics.InventoryCheckDuration.Observe(time.Since(checkStart).Seconds())
			metrics.OrdersProcessed.WithLabelValues("failure").Inc()
			metrics.OrderProcessingDuration.Observe(time.Since(start).Seconds())
			return fmt.Errorf("deduct stock for product %s: %w", item.ProductID, err)
		}
		metrics.InventoryCheckDuration.Observe(time.Since(checkStart).Seconds())
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

	// Step 3: Simulate payment processing (2-4 seconds)
	metrics.PaymentAttempts.WithLabelValues("initiated").Inc()
	paymentStart := time.Now()

	randomDuration, err := rand.Int(rand.Reader, big.NewInt(3))
	if err != nil {
		metrics.PaymentAttempts.WithLabelValues("failure").Inc()
		metrics.OrdersProcessed.WithLabelValues("failure").Inc()
		metrics.OrderProcessingDuration.Observe(time.Since(start).Seconds())
		return fmt.Errorf("generate random duration: %w", err)
	}
	sleepTime := time.Duration(randomDuration.Int64()+2) * time.Second
	time.Sleep(sleepTime)

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
