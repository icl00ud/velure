package queue

import (
	"encoding/json"
	"errors"
	"fmt"
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

func NewRabbitMQRepo() (*RabbitMQRepository, error) {
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
		return nil, fmt.Errorf("%w: %s", ErrMissingEnvVar, strings.Join(missing, ", "))
	}

	amqpURL := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		os.Getenv("RABBITMQ_DEFAULT_USER"),
		os.Getenv("RABBITMQ_DEFAULT_PASS"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)

	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
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
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	fmt.Println("Connected to RabbitMQ successfully")

	return &RabbitMQRepository{
		conn:     conn,
		channel:  ch,
		exchange: exchangeName,
	}, nil
}

func (r *RabbitMQRepository) PublishEvent(event domain.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
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
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

func (r *RabbitMQRepository) Close() error {
	var errs []error

	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("channel close error: %w", err))
		}
	}

	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("connection close error: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("rabbitmq shutdown errors: %v", errs)
	}

	return nil
}
