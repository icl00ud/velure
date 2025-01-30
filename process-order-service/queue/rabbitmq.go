package queue

import (
	"errors"
	"fmt"
	"os"
	"strings"

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
		"RABBITMQ_USER",
		"RABBITMQ_PASS",
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
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASS"),
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

	return &RabbitMQRepository{
		conn:     conn,
		channel:  ch,
		exchange: exchangeName,
	}, nil
}

func (r *RabbitMQRepository) DeclareQueue(queueName string) (amqp091.Queue, error) {
	return r.channel.QueueDeclare(
		queueName,
		true,  // Durable
		false, // Delete when unused
		false, // Exclusive
		false, // No-wait
		nil,   // Arguments
	)
}

func (r *RabbitMQRepository) BindQueue(queueName, routingKey string) error {
	return r.channel.QueueBind(
		queueName,
		routingKey,
		r.exchange,
		false,
		nil,
	)
}

func (r *RabbitMQRepository) Consume(queueName string, consumerTag string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error) {
	return r.channel.Consume(
		queueName,
		consumerTag,
		autoAck,
		exclusive,
		noLocal,
		noWait,
		args,
	)
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

func (r *RabbitMQRepository) GetExchangeName() string {
	return r.exchange
}
