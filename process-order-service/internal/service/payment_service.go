package service

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/icl00ud/process-order-service/internal/model"
	"github.com/icl00ud/process-order-service/internal/queue"
)

type PaymentService interface {
	Process(orderID string, amount int) error
}

type paymentService struct {
	pub queue.Publisher
}

func NewPaymentService(pub queue.Publisher) PaymentService {
	return &paymentService{pub: pub}
}

func (s *paymentService) Process(orderID string, amount int) error {
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

	sleepTime := time.Duration(rand.Uint32()) % 3 * time.Second
	fmt.Printf("Sleeping for %v", sleepTime)
	time.Sleep(sleepTime)

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
