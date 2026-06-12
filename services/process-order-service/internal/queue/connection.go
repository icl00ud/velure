package queue

import (
	"context"
	"fmt"

	"github.com/icl00ud/velure/shared/logger"
)

type RabbitMQConnection struct {
	conn   AMQPConnection
	url    string
	logger *logger.Logger
}

func NewRabbitMQConnection(amqpURL string, log *logger.Logger) (*RabbitMQConnection, error) {
	conn, err := amqpDial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	log.Info("rabbitmq connection established")
	return &RabbitMQConnection{conn: conn, url: amqpURL, logger: log}, nil
}

func (r *RabbitMQConnection) NewConsumer(queueName string) (Consumer, error) {
	ch, err := r.consumerChannel(r.conn, queueName)
	if err != nil {
		return nil, err
	}

	c := &rabbitMQConsumer{conn: nil, channel: ch, queue: queueName, logger: r.logger}
	// A broker restart closes the shared connection and the deliveries
	// channel with it; redial on a dedicated connection so the consumer
	// survives the outage instead of silently going idle.
	c.reconnect = func(ctx context.Context) (AMQPChannel, error) {
		conn, err := amqpDial(r.url)
		if err != nil {
			return nil, fmt.Errorf("redial rabbitmq: %w", err)
		}
		ch, err := r.consumerChannel(conn, queueName)
		if err != nil {
			conn.Close()
			return nil, err
		}
		if c.conn != nil {
			_ = c.conn.Close()
		}
		c.conn = conn
		return ch, nil
	}
	return c, nil
}

// consumerChannel opens and configures a channel for queue consumption.
// The queue itself is created by RabbitMQ's bootstrap.sh with specific
// arguments (DLX, etc); redeclaring it here would fail with
// PRECONDITION_FAILED.
func (r *RabbitMQConnection) consumerChannel(conn AMQPConnection, queueName string) (AMQPChannel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Bind queue to exchange with routing key pattern for order events
	// (matches order.created, order.completed, etc.)
	if err := ch.QueueBind(queueName, "order.*", "orders", false, nil); err != nil {
		ch.Close()
		return nil, fmt.Errorf("queue bind: %w", err)
	}

	if err := ch.Qos(1, 0, false); err != nil {
		ch.Close()
		return nil, err
	}
	return ch, nil
}

func (r *RabbitMQConnection) NewPublisher(exchange string) (Publisher, error) {
	ch, err := r.publisherChannel(r.conn, exchange)
	if err != nil {
		return nil, err
	}

	p := &rabbitPublisher{conn: nil, channel: ch, exchange: exchange, logger: r.logger}
	// Redial on a dedicated connection so status events survive a broker
	// restart instead of being silently dropped.
	p.reconnect = func() (AMQPChannel, error) {
		conn, err := amqpDial(r.url)
		if err != nil {
			return nil, fmt.Errorf("redial rabbitmq: %w", err)
		}
		ch, err := r.publisherChannel(conn, exchange)
		if err != nil {
			conn.Close()
			return nil, err
		}
		if p.conn != nil {
			_ = p.conn.Close()
		}
		p.conn = conn
		return ch, nil
	}
	return p, nil
}

func (r *RabbitMQConnection) publisherChannel(conn AMQPConnection, exchange string) (AMQPChannel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		ch.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}
	return ch, nil
}

func (r *RabbitMQConnection) Close() error {
	return r.conn.Close()
}
