package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/icl00ud/publish-order-service/domain"
	"github.com/rabbitmq/amqp091-go"
)

var (
	ErrMissingEnvVar = errors.New("missing environment variables")
)

type RabbitMQRepository struct {
	conn     *amqp091.Connection
	channel  *amqp091.Channel
	exchange string
}

func NewRabbitMQRepo() *RabbitMQRepository {
	requiredVars := []string{
		"RABBITMQ_HOST",
		"RABBITMQ_PORT",
		"RABBITMQ_DEFAULT_USER",
		"RABBITMQ_DEFAULT_PASS",
		"RABBITMQ_EXCHANGE",
	}

	missing := make([]string, 0)
	for _, varName := range requiredVars {
		if os.Getenv(varName) == "" {
			missing = append(missing, varName)
		}
	}
	if len(missing) > 0 {
		log.Fatalf("%v: %s", ErrMissingEnvVar, strings.Join(missing, ", "))
	}

	amqpURL := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		os.Getenv("RABBITMQ_DEFAULT_USER"),
		os.Getenv("RABBITMQ_DEFAULT_PASS"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)

	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		log.Fatalf("failed to open channel: %v", err)
	}

	exchangeName := os.Getenv("RABBITMQ_EXCHANGE")
	err = ch.ExchangeDeclare(
		exchangeName,
		"topic", // Tipo de exchange
		true,    // Durable
		false,   // Auto-deleted
		false,   // Internal
		false,   // No-wait
		nil,     // Arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		log.Fatalf("failed to declare exchange: %v", err)
	}

	log.Println("Connected to RabbitMQ successfully")

	return &RabbitMQRepository{
		conn:     conn,
		channel:  ch,
		exchange: exchangeName,
	}
}

func (r *RabbitMQRepository) PublishEvent(event domain.Event) {
	body, err := json.Marshal(event)
	if err != nil {
		log.Fatalf("failed to marshal event: %v", err)
	}

	err = r.channel.Publish(
		r.exchange,         // Exchange
		string(event.Type), // Routing key
		false,              // Mandatory
		false,              // Immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		log.Fatalf("failed to publish event: %v", err)
	}
}

func (r *RabbitMQRepository) Close() {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			log.Printf("channel close error: %v", err)
		}
	}

	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			log.Printf("connection close error: %v", err)
		}
	}
}
