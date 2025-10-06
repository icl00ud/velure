package publisher

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/icl00ud/publish-order-service/internal/model"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Publisher interface {
	Publish(evt model.Event) error
	Close() error
}

type rabbitMQPublisher struct {
	conn     *amqp091.Connection
	ch       *amqp091.Channel
	exchange string
	logger   *zap.Logger
	closed   bool
	mu       sync.Mutex
}

func NewRabbitMQPublisher(amqpURL string, exchange string, logger *zap.Logger) (Publisher, error) {
	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}
	logger.Info("RabbitMQ dial succeeded", zap.String("url", amqpURL))

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}
	logger.Info("RabbitMQ channel opened")

	if err := ch.ExchangeDeclare(
		exchange,
		"topic",
		true,  // durable
		false, // autoDelete
		false, // internal
		false, // noWait
		nil,   // args
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}
	logger.Info("Exchange declared", zap.String("exchange", exchange))

	return &rabbitMQPublisher{
		conn:     conn,
		ch:       ch,
		exchange: exchange,
		logger:   logger,
	}, nil
}

func (r *rabbitMQPublisher) Publish(evt model.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return fmt.Errorf("publisher is closed")
	}

	body, err := json.Marshal(evt)
	if err != nil {
		r.logger.Error("failed to marshal event", zap.Error(err), zap.Any("event", evt))
		return err
	}
	err = r.ch.Publish(
		r.exchange,
		evt.Type,
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		r.logger.Error("publish failed", zap.Error(err), zap.String("exchange", r.exchange), zap.String("routingKey", evt.Type))
		return err
	}
	r.logger.Info("event published successfully", zap.String("exchange", r.exchange), zap.String("routingKey", evt.Type), zap.Int("body_size", len(body)))
	return nil
}

func (r *rabbitMQPublisher) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}
	r.closed = true

	if r.ch != nil {
		if err := r.ch.Close(); err != nil {
			r.logger.Warn("channel close error", zap.Error(err))
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			r.logger.Warn("connection close error", zap.Error(err))
			return err
		}
	}
	r.logger.Info("RabbitMQ connection closed")
	return nil
}
