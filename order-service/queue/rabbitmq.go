package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/icl00ud/velure-order-service/domain"
	"github.com/rabbitmq/amqp091-go"
)

var (
	ErrMissingEnvVar = errors.New("missing environment variables")
	ErrInvalidEnvVar = errors.New("invalid environment variables")
)

type RabbitMQRepository struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	queue   string
}

// NewRabbitMQRepo creates a new RabbitMQ repository with connection and channel
func NewRabbitMQRepo() (*RabbitMQRepository, error) {
	requiredVars := []string{
		"RABBITMQ_HOST",
		"RABBITMQ_PORT",
		"RABBITMQ_USER",
		"RABBITMQ_PASS",
		"RABBITMQ_QUEUE",
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
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASS"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)

	// Establish connection
	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare durable queue
	queueName := os.Getenv("RABBITMQ_QUEUE")
	_, err = ch.QueueDeclare(
		queueName,
		true,  // Durable
		false, // Delete when unused
		false, // Exclusive
		false, // No-wait
		nil,   // Arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &RabbitMQRepository{
		conn:    conn,
		channel: ch,
		queue:   queueName,
	}, nil
}

// PublishOrder publishes an order to the queue with persistent delivery
func (r *RabbitMQRepository) PublishOrder(order domain.Order) error {
	body, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	err = r.channel.Publish(
		"",      // Exchange
		r.queue, // Routing key
		false,   // Mandatory
		false,   // Immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent, // Persistent message
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Close safely closes all connections and collects errors
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
