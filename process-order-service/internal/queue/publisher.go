package queue

import (
	"encoding/json"
	"fmt"

	"github.com/icl00ud/process-order-service/internal/model"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Publisher interface {
	Publish(evt model.Event) error
	Close() error
}

type rabbitPublisher struct {
	conn     *amqp091.Connection
	channel  *amqp091.Channel
	exchange string
	logger   *zap.Logger
}

func NewRabbitPublisher(amqpURL, exchange string, logger *zap.Logger) (Publisher, error) {
	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}
	if err := ch.ExchangeDeclare(
		exchange,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}
	logger.Info("Publisher ready", zap.String("exchange", exchange))
	return &rabbitPublisher{conn: conn, channel: ch, exchange: exchange, logger: logger}, nil
}

func (r *rabbitPublisher) Publish(evt model.Event) error {
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	if err := r.channel.Publish(
		r.exchange,
		evt.Type,
		false, false,
		amqp091.Publishing{ContentType: "application/json", Body: body},
	); err != nil {
		return err
	}
	r.logger.Debug("event published", zap.String("type", evt.Type))
	return nil
}

func (r *rabbitPublisher) Close() error {
	r.channel.Close()
	return r.conn.Close()
}
