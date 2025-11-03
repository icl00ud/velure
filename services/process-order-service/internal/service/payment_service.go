package service

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/icl00ud/process-order-service/internal/client"
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
	// Step 1: Deduct stock for all items BEFORE processing payment
	for _, item := range items {
		if err := s.productClient.UpdateQuantity(item.ProductID, -item.Quantity); err != nil {
			return fmt.Errorf("deduct stock for product %s: %w", item.ProductID, err)
		}
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
	randomDuration, err := rand.Int(rand.Reader, big.NewInt(3))
	if err != nil {
		return fmt.Errorf("generate random duration: %w", err)
	}
	sleepTime := time.Duration(randomDuration.Int64()+2) * time.Second
	time.Sleep(sleepTime)

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
		return fmt.Errorf("publish completed: %w", err)
	}
	return nil
}
