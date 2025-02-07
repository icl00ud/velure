package queue

import (
	"fmt"
	"log"
	"os"

	"github.com/icl00ud/process-order-service/domain"
	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	Conn    *amqp091.Connection
	Channel *amqp091.Channel
	Queue   string
}

func NewConsumer() *Consumer {
	requiredVars := []string{
		"RABBITMQ_HOST",
		"RABBITMQ_PORT",
		"RABBITMQ_USER",
		"RABBITMQ_PASS",
		"RABBITMQ_EXCHANGE",
		"RABBITMQ_QUEUE",
	}

	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			log.Fatalf("Missing environment variable: %s", v)
		}
	}

	amqpURL := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASS"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)

	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open channel: %v", err)
	}

	err = ch.Qos(50, 0, false)
	if err != nil {
		log.Fatalf("failed to set QoS: %v", err)
	}

	exchangeName := os.Getenv("RABBITMQ_EXCHANGE")
	queueName := os.Getenv("RABBITMQ_QUEUE")
	// Declara a fila
	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatalf("failed to declare queue: %v", err)
	}
	// Associa a fila à exchange com o routing key "order.created"
	err = ch.QueueBind(
		q.Name,
		string(domain.OrderCreated),
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed to bind queue: %v", err)
	}

	return &Consumer{
		Conn:    conn,
		Channel: ch,
		Queue:   q.Name,
	}
}

func (c *Consumer) Consume() (<-chan amqp091.Delivery, error) {
	return c.Channel.Consume(
		c.Queue,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
}

func (c *Consumer) Close() {
	if c.Channel != nil {
		if err := c.Channel.Close(); err != nil {
			log.Printf("channel close error: %v", err)
		}
	}
	if c.Conn != nil {
		if err := c.Conn.Close(); err != nil {
			log.Printf("connection close error: %v", err)
		}
	}
}
