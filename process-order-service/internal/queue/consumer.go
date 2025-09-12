package queue

import (
	"context"
	"encoding/json"

	"github.com/icl00ud/process-order-service/internal/model"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Consumer interface {
	Consume(ctx context.Context, handler func(model.Event) error) error
	Close() error
}

type rabbitMQConsumer struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	queue   string
	logger  *zap.Logger
}

func NewRabbitMQConsumer(amqpURL, queueName string, logger *zap.Logger) (Consumer, error) {
	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &rabbitMQConsumer{conn: conn, channel: ch, queue: queueName, logger: logger}, nil
}

func (r *rabbitMQConsumer) Consume(ctx context.Context, handler func(model.Event) error) error {
	msgs, err := r.channel.Consume(r.queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d := <-msgs:
			var evt model.Event
			if err := json.Unmarshal(d.Body, &evt); err != nil {
				d.Nack(false, true)
				r.logger.Error("invalid event", zap.Error(err))
				continue
			}
			if err := handler(evt); err != nil {
				d.Nack(false, true)
				r.logger.Error("handler failed", zap.Error(err))
				continue
			}
			d.Ack(false)
		}
	}
}

func (r *rabbitMQConsumer) Close() error {
	r.channel.Close()
	return r.conn.Close()
}
