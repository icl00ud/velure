package publisher

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/icl00ud/publish-order-service/internal/model"
	"github.com/icl00ud/velure-shared/logger"
	"github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	Publish(evt model.Event) error
	Close() error
}

type rabbitMQPublisher struct {
	amqpURL   string
	conn      amqpPublisherConn
	ch        amqpPublisherChannel
	exchange  string
	logger    *logger.Logger
	closed    bool
	mu        sync.Mutex
	dialFn    func(string) (amqpPublisherConn, error)
	connectFn func() error
}

type amqpPublisherConn interface {
	Channel() (amqpPublisherChannel, error)
	Close() error
}

type amqpPublisherChannel interface {
	Publish(exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp091.Table) error
	Close() error
}

var amqpDialer = amqp091.Dial

var dialPublisher = func(amqpURL string) (amqpPublisherConn, error) {
	conn, err := amqpDialer(amqpURL)
	if err != nil {
		return nil, err
	}
	return &livePublisherConn{conn: conn}, nil
}

func NewRabbitMQPublisher(amqpURL string, exchange string, log *logger.Logger) (Publisher, error) {
	return newRabbitMQPublisher(amqpURL, exchange, log, dialPublisher)
}

func newRabbitMQPublisher(amqpURL string, exchange string, log *logger.Logger, dialFn func(string) (amqpPublisherConn, error)) (Publisher, error) {
	p := &rabbitMQPublisher{
		amqpURL:  amqpURL,
		exchange: exchange,
		logger:   log,
		dialFn:   dialFn,
	}
	p.connectFn = p.connect

	if err := p.connectFn(); err != nil {
		return nil, err
	}

	return p, nil
}

type livePublisherConn struct {
	conn dialConnection
}

func (c *livePublisherConn) Channel() (amqpPublisherChannel, error) {
	if c.conn == nil {
		return nil, amqp091.ErrClosed
	}
	return c.conn.Channel()
}

func (c *livePublisherConn) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

type dialConnection interface {
	Channel() (*amqp091.Channel, error)
	Close() error
}

func (r *rabbitMQPublisher) connect() error {
	// Close existing connection/channel if any, ignoring errors
	if r.ch != nil {
		_ = r.ch.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}

	conn, err := r.dialFn(r.amqpURL)
	if err != nil {
		return fmt.Errorf("dial rabbitmq: %w", err)
	}
	r.logger.Info("RabbitMQ dial succeeded")

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("open channel: %w", err)
	}
	r.logger.Info("RabbitMQ channel opened")

	if err := ch.ExchangeDeclare(
		r.exchange,
		"topic",
		true,  // durable
		false, // autoDelete
		false, // internal
		false, // noWait
		nil,   // args
	); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("declare exchange: %w", err)
	}
	r.logger.Info("Exchange declared", logger.String("exchange", r.exchange))

	r.conn = conn
	r.ch = ch
	return nil
}

func (r *rabbitMQPublisher) Publish(evt model.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return fmt.Errorf("publisher is closed")
	}

	body, err := json.Marshal(evt)
	if err != nil {
		r.logger.Error("failed to marshal event", logger.Err(err), logger.Any("event", evt))
		return err
	}

	publishFunc := func() error {
		if r.ch == nil {
			return amqp091.ErrClosed
		}
		return r.ch.Publish(
			r.exchange,
			evt.Type,
			false, // mandatory
			false, // immediate
			amqp091.Publishing{
				ContentType: "application/json",
				Body:        body,
			},
		)
	}

	err = publishFunc()
	if err != nil {
		r.logger.Warn("publish failed, attempting reconnect", logger.Err(err))
		if recErr := r.connectFn(); recErr != nil {
			r.logger.Error("reconnect failed", logger.Err(recErr))
			return err
		}
		if err = publishFunc(); err != nil {
			r.logger.Error("publish failed after reconnect", logger.Err(err))
			return err
		}
	}

	r.logger.Info("event published", logger.String("exchange", r.exchange), logger.String("routingKey", evt.Type), logger.Int("body_size", len(body)))
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
			r.logger.Warn("channel close error", logger.Err(err))
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			r.logger.Warn("connection close error", logger.Err(err))
			return err
		}
	}
	r.logger.Info("RabbitMQ connection closed")
	return nil
}
