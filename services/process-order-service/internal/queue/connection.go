package queue

import (
	"fmt"

	"go.uber.org/zap"
)

type RabbitMQConnection struct {
	conn   AMQPConnection
	logger *zap.Logger
}

func NewRabbitMQConnection(amqpURL string, logger *zap.Logger) (*RabbitMQConnection, error) {
	conn, err := amqpDial(amqpURL)
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

	// Não redeclarar a fila aqui - ela é criada pelo bootstrap.sh do RabbitMQ
	// com argumentos específicos (DLX, etc). Redeclarar causaria PRECONDITION_FAILED.
	// _, err = ch.QueueDeclare(queueName, true, false, false, false, nil)

	// Bind queue to exchange with routing key pattern for order events
	err = ch.QueueBind(
		queueName, // queue name
		"order.*", // routing key pattern (matches order.created, order.completed, etc.)
		"orders",  // exchange name
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		ch.Close()
		return nil, fmt.Errorf("queue bind: %w", err)
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
