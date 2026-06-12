package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/icl00ud/velure/shared/logger"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/icl00ud/velure/services/publish-order-service/internal/model"
	"github.com/icl00ud/velure/services/publish-order-service/internal/telemetry"
)

const (
	dlxExchange = "publish.dlx"
	maxRetries  = 3
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

	// reconnect re-establishes a channel after the broker drops the
	// connection. When nil, Start gives up once the deliveries channel
	// closes (legacy behavior, used by tests without a broker).
	reconnect      func(ctx context.Context) (amqpChan, error)
	reconnectDelay time.Duration
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
	conn, ch, q, err := setupConsumerChannel(amqpURL, exchange, queueName)
	if err != nil {
		return nil, err
	}

	log.Info("rabbitmq consumer initialized",
		logger.String("exchange", exchange),
		logger.String("queue", queueName),
		logger.Int("workers", workers))

	c := &rabbitConsumer{
		conn:    conn,
		channel: ch,
		queue:   q,
		handler: handler,
		logger:  log,
		workers: workers,
	}
	// A broker restart closes the deliveries channel; redial with a full
	// re-setup so the consumer survives the outage instead of going idle.
	c.reconnect = func(ctx context.Context) (amqpChan, error) {
		conn, ch, _, err := setupConsumerChannel(amqpURL, exchange, queueName)
		if err != nil {
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

// setupConsumerChannel dials the broker and declares the exchange, queue,
// bindings and QoS the status consumer depends on.
func setupConsumerChannel(amqpURL, exchange, queueName string) (amqpConn, amqpChan, string, error) {
	conn, err := dialRabbitMQ(amqpURL)
	if err != nil {
		return nil, nil, "", fmt.Errorf("dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, "", fmt.Errorf("open channel: %w", err)
	}

	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, nil, "", fmt.Errorf("declare exchange: %w", err)
	}

	q, err := ch.QueueDeclare(queueName, true, false, false, false, amqp091.Table{
		"x-dead-letter-exchange": dlxExchange,
		"x-max-length":           int32(10000),
	})
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, nil, "", fmt.Errorf("declare queue: %w", err)
	}

	for _, key := range []string{"order.processing", "order.completed", "order.failed"} {
		if err := ch.QueueBind(q.Name, key, exchange, false, nil); err != nil {
			ch.Close()
			conn.Close()
			return nil, nil, "", fmt.Errorf("bind queue to %s: %w", key, err)
		}
	}

	if err := ch.Qos(1, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, nil, "", fmt.Errorf("set qos: %w", err)
	}
	return conn, ch, q.Name, nil
}

func (r *rabbitConsumer) Start(ctx context.Context) error {
	for {
		msgs, err := r.channel.Consume(r.queue, "", false, false, false, false, nil)
		if err != nil {
			if r.reconnect == nil {
				return fmt.Errorf("start consuming: %w", err)
			}
			r.logger.Warn("consume failed, reconnecting", logger.Err(err))
			if err := r.redial(ctx); err != nil {
				return err
			}
			continue
		}

		var wg sync.WaitGroup
		for i := 0; i < r.workers; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				r.worker(ctx, id, msgs)
			}(i)
		}
		// Workers exit either on ctx cancellation or when the broker closes
		// the deliveries channel; only the latter warrants a reconnect.
		wg.Wait()
		if ctx.Err() != nil {
			return nil
		}
		if r.reconnect == nil {
			return nil
		}
		r.logger.Warn("deliveries channel closed, reconnecting")
		if err := r.redial(ctx); err != nil {
			return err
		}
	}
}

// redial retries r.reconnect with a fixed delay until it succeeds or the
// context is cancelled.
func (r *rabbitConsumer) redial(ctx context.Context) error {
	delay := r.reconnectDelay
	if delay == 0 {
		delay = 5 * time.Second
	}
	for {
		ch, err := r.reconnect(ctx)
		if err == nil {
			r.channel = ch
			r.logger.Info("rabbitmq consumer reconnected")
			return nil
		}
		r.logger.Error("reconnect failed, retrying", logger.Err(err))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
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
				retryCount := getRetryCount(msg.Headers)
				if retryCount >= maxRetries {
					r.logger.Error("max retries exceeded - sending to DLQ",
						logger.Int("worker_id", id),
						logger.Int64("retry_count", retryCount),
						logger.Err(err))
					msg.Nack(false, false)
				} else {
					r.logger.Warn("transient error - requeueing for retry",
						logger.Int("worker_id", id),
						logger.Int64("retry_count", retryCount),
						logger.Err(err))
					msg.Nack(false, true)
				}
			} else {
				msg.Ack(false)
			}
		}
	}
}

// headersToMap converts string-valued AMQP headers (where the W3C trace
// context travels) into the map form the OTel propagator understands.
func headersToMap(headers amqp091.Table) map[string]string {
	m := make(map[string]string, len(headers))
	for k, v := range headers {
		if s, ok := v.(string); ok {
			m[k] = s
		}
	}
	return m
}

func getRetryCount(headers amqp091.Table) int64 {
	if headers == nil {
		return 0
	}
	xDeath, ok := headers["x-death"].([]any)
	if !ok || len(xDeath) == 0 {
		return 0
	}
	death, ok := xDeath[0].(amqp091.Table)
	if !ok {
		return 0
	}
	count, ok := death["count"].(int64)
	if !ok {
		return 0
	}
	return count
}

func (r *rabbitConsumer) processMessage(ctx context.Context, msg amqp091.Delivery) error {
	var evt model.Event
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		r.logger.Error("failed to unmarshal event", logger.Err(err))
		return err
	}

	// Continue the trace propagated by the producer through AMQP headers.
	ctx = telemetry.ExtractMap(ctx, headersToMap(msg.Headers))
	ctx, span := otel.Tracer("status-consumer").Start(ctx, "consume "+evt.Type,
		trace.WithSpanKind(trace.SpanKindConsumer))
	defer span.End()

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
