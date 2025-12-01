package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/icl00ud/velure-shared/logger"
	"github.com/rabbitmq/amqp091-go"

	"github.com/icl00ud/publish-order-service/internal/model"
)

type EventHandler func(ctx context.Context, evt model.Event) error

type Consumer interface {
	Start(ctx context.Context) error
	Close() error
}

type amqpConn interface {
	Channel() (amqpChan, error)
	Close() error
}

type amqpChan interface {
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error)
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp091.Table) error
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp091.Table) (amqp091.Queue, error)
	QueueBind(name, key, exchange string, noWait bool, args amqp091.Table) error
	Qos(prefetchCount, prefetchSize int, global bool) error
	Close() error
}

type rabbitConsumer struct {
	conn    amqpConn
	channel amqpChan
	queue   string
	handler EventHandler
	logger  *logger.Logger
	workers int
}

type liveConsumerConn struct {
	conn *amqp091.Connection
}

func (c *liveConsumerConn) Channel() (amqpChan, error) {
	if c.conn == nil {
		return nil, amqp091.ErrClosed
	}
	return c.conn.Channel()
}

func (c *liveConsumerConn) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

var dialRabbitMQ = func(amqpURL string) (amqpConn, error) {
	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, err
	}
	return &liveConsumerConn{conn: conn}, nil
}

func NewRabbitMQConsumer(amqpURL, exchange, queueName string, handler EventHandler, workers int, log *logger.Logger) (Consumer, error) {
	conn, err := dialRabbitMQ(amqpURL)
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

	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	if err := ch.QueueBind(q.Name, "order.processing", exchange, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("bind queue to order.processing: %w", err)
	}

	if err := ch.QueueBind(q.Name, "order.completed", exchange, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("bind queue to order.completed: %w", err)
	}

	if err := ch.Qos(1, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("set qos: %w", err)
	}

	log.Info("rabbitmq consumer initialized",
		logger.String("exchange", exchange),
		logger.String("queue", queueName),
		logger.Int("workers", workers))

	return &rabbitConsumer{
		conn:    conn,
		channel: ch,
		queue:   q.Name,
		handler: handler,
		logger:  log,
		workers: workers,
	}, nil
}

func (r *rabbitConsumer) Start(ctx context.Context) error {
	msgs, err := r.channel.Consume(r.queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("start consuming: %w", err)
	}

	for i := 0; i < r.workers; i++ {
		go r.worker(ctx, i, msgs)
	}

	<-ctx.Done()
	return nil
}

func (r *rabbitConsumer) worker(ctx context.Context, id int, msgs <-chan amqp091.Delivery) {
	r.logger.Info("consumer worker started", logger.Int("worker_id", id))

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("consumer worker stopped", logger.Int("worker_id", id))
			return
		case msg, ok := <-msgs:
			if !ok {
				r.logger.Warn("message channel closed", logger.Int("worker_id", id))
				return
			}

			if err := r.processMessage(ctx, msg); err != nil {
				r.logger.Error("message processing failed",
					logger.Int("worker_id", id),
					logger.Err(err))
				msg.Nack(false, true)
			} else {
				msg.Ack(false)
			}
		}
	}
}

func (r *rabbitConsumer) processMessage(ctx context.Context, msg amqp091.Delivery) error {
	var evt model.Event
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		r.logger.Error("failed to unmarshal event", logger.Err(err))
		return err
	}

	r.logger.Info("processing event",
		logger.String("type", evt.Type),
		logger.String("payload", string(evt.Payload)))

	if err := r.handler(ctx, evt); err != nil {
		r.logger.Error("handler failed",
			logger.String("event_type", evt.Type),
			logger.Err(err))
		return err
	}

	r.logger.Info("event processed", logger.String("type", evt.Type))
	return nil
}

func (r *rabbitConsumer) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
