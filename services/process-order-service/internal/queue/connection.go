package queue

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type RabbitMQConnection struct {
	conn   *amqp091.Connection
	logger *zap.Logger
}

func NewRabbitMQConnection(amqpURL string, logger *zap.Logger) (*RabbitMQConnection, error) {
	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	logger.Info("rabbitmq connection established")
	return &RabbitMQConnection{conn: conn, logger: logger}, nil
}

func (r *RabbitMQConnection) NewConsumer(queueName string) (Consumer, error) {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		return nil, err
	}

	if err := ch.Qos(1, 0, false); err != nil {
		ch.Close()
		return nil, err
	}

	return &rabbitMQConsumer{conn: nil, channel: ch, queue: queueName, logger: r.logger}, nil
}

func (r *RabbitMQConnection) NewPublisher(exchange string) (Publisher, error) {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}

	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		ch.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	return &rabbitPublisher{conn: nil, channel: ch, exchange: exchange, logger: r.logger}, nil
}

func (r *RabbitMQConnection) Close() error {
	return r.conn.Close()
}
