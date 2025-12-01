package queue

import (
	"encoding/json"
	"fmt"

	"github.com/icl00ud/process-order-service/internal/model"
	"github.com/icl00ud/velure-shared/logger"
	"github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	Publish(evt model.Event) error
	Close() error
}

type rabbitPublisher struct {
	conn     AMQPConnection
	channel  AMQPChannel
	exchange string
	logger   *logger.Logger
}

func NewRabbitPublisher(amqpURL, exchange string, log *logger.Logger) (Publisher, error) {
	conn, err := amqpDial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	return &rabbitPublisher{conn: conn, channel: ch, exchange: exchange, logger: log}, nil
}

func (r *rabbitPublisher) Publish(evt model.Event) error {
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	err = r.channel.Publish(
		r.exchange,
		evt.Type,
		false, false,
		amqp091.Publishing{ContentType: "application/json", Body: body},
	)
	if err != nil {
		r.logger.Error("publish failed", logger.Err(err), logger.String("exchange", r.exchange), logger.String("event_type", evt.Type))
		return err
	}
	r.logger.Info("payment event published", logger.String("exchange", r.exchange), logger.String("event_type", evt.Type))
	return nil
}

func (r *rabbitPublisher) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
