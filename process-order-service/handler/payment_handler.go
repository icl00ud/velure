package handler

import (
	"encoding/json"
	"log"

	"github.com/icl00ud/process-order-service/domain"
	"github.com/icl00ud/process-order-service/queue"
)

type OrderConsumer struct {
	repo *queue.RabbitMQRepository
}

func NewOrderConsumer(repo *queue.RabbitMQRepository) *OrderConsumer {
	return &OrderConsumer{
		repo: repo,
	}
}

func (c *OrderConsumer) StartConsuming() {
	queueName := "order_created_queue"
	routingKey := string(domain.OrderCreated)

	q, err := c.repo.DeclareQueue(queueName)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	err = c.repo.BindQueue(q.Name, routingKey)
	if err != nil {
		log.Fatalf("Failed to bind queue: %v", err)
	}

	log.Printf("Queue %s bound to exchange %s with routing key %s", q.Name, c.repo.GetExchangeName(), routingKey)

	msgs, err := c.repo.Consume(
		q.Name, // Nome da fila
		"",     // Consumer Tag
		true,   // Auto-ack
		false,  // Exclusive
		false,  // No-local
		false,  // No-wait
		nil,    // Args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	log.Println("Iniciando o consumo de mensagens...")

	// Processa as mensagens recebidas
	go func() {
		for d := range msgs {
			var event domain.Event
			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.Printf("Failed to unmarshal event: %v", err)
				continue
			}

			// Processa o evento com base no tipo
			switch event.Type {
			case domain.OrderCreated:
				handleOrderCreated(event.Order)
			// Adicione outros casos conforme necessário
			default:
				log.Printf("Unknown event type: %s", event.Type)
			}
		}
	}()
}

func handleOrderCreated(order domain.Order) {
	log.Printf("Processando OrderCreated para o pedido ID: %s, Amount: %d", order.ID, order.Amount)
}
