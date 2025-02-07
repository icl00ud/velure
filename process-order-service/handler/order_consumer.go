package handler

import (
	"encoding/json"
	"log"
	"time"

	"github.com/icl00ud/process-order-service/domain"
	"github.com/icl00ud/process-order-service/queue"
	"github.com/icl00ud/process-order-service/storage"
)

type OrderConsumer struct {
	consumer *queue.Consumer
}

func NewOrderConsumer(consumer *queue.Consumer) *OrderConsumer {
	return &OrderConsumer{
		consumer: consumer,
	}
}

func (oc *OrderConsumer) StartConsuming(ps *storage.PaymentStorage) {
	msgs, err := oc.consumer.Consume()
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	workerCount := 20
	for i := 0; i < workerCount; i++ {
		go func(workerID int) {
			for d := range msgs {
				var event domain.Event
				if err := json.Unmarshal(d.Body, &event); err != nil {
					d.Nack(false, true)
					continue
				}

				if event.Type == domain.OrderCreated {
					var order struct {
						ID    string `json:"id"`
						Total int    `json:"total"`
					}
					if err := json.Unmarshal(event.Payload, &order); err != nil {
						d.Nack(false, true)
						continue
					}

					payment := domain.Payment{
						ID:          order.ID,
						OrderID:     order.ID,
						Amount:      order.Total,
						Status:      domain.PaymentProcessed,
						ProcessedAt: time.Now(),
					}

					ps.StorePayment(payment)
				}

				d.Ack(false)
			}
		}(i)
	}

	log.Println("Waiting for messages")
	select {}
}
