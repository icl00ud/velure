package queue

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/icl00ud/velure/services/process-order-service/internal/client"
	"github.com/icl00ud/velure/services/process-order-service/internal/model"
	"github.com/icl00ud/velure/shared/logger"
	"github.com/rabbitmq/amqp091-go"
)

type stubChannel struct {
	deliveries chan amqp091.Delivery
	consumeErr error
	qosCalled  bool
	closed     bool
}

func (s *stubChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error) {
	if s.consumeErr != nil {
		return nil, s.consumeErr
	}
	return s.deliveries, nil
}

func (s *stubChannel) Qos(prefetchCount, prefetchSize int, global bool) error {
	s.qosCalled = true
	return nil
}

func (s *stubChannel) QueueBind(queue, key, exchange string, noWait bool, args amqp091.Table) error {
	return nil
}

func (s *stubChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
	return nil
}

func (s *stubChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp091.Table) error {
	return nil
}

func (s *stubChannel) Close() error {
	s.closed = true
	close(s.deliveries)
	return nil
}

type stubConn struct {
	closed bool
}

func (s *stubConn) Channel() (AMQPChannel, error) { return nil, nil }
func (s *stubConn) Close() error {
	s.closed = true
	return nil
}

type stubAcker struct {
	acked    bool
	nacked   bool
	requeue  bool
	multiple bool
}

func (s *stubAcker) Ack(tag uint64, multiple bool) error {
	s.acked = true
	s.multiple = multiple
	return nil
}

func (s *stubAcker) Nack(tag uint64, multiple bool, requeue bool) error {
	s.nacked = true
	s.multiple = multiple
	s.requeue = requeue
	return nil
}

func (s *stubAcker) Reject(tag uint64, requeue bool) error {
	s.nacked = true
	s.requeue = requeue
	return nil
}

func TestRabbitMQConsumer_ConsumeSuccess(t *testing.T) {
	acker := &stubAcker{}
	deliveries := make(chan amqp091.Delivery, 1)
	deliveries <- amqp091.Delivery{Body: []byte(`{"type":"order.created","payload":{}}`), Acknowledger: acker}
	close(deliveries)

	ch := &stubChannel{deliveries: deliveries}
	c := &rabbitMQConsumer{
		conn:    &stubConn{},
		channel: ch,
		queue:   "q",
		logger:  logger.NewNop(),
	}

	if err := c.Consume(context.Background(), func(_ string, evt model.Event) error { return nil }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !acker.acked {
		t.Fatal("expected ack")
	}
}

func TestRabbitMQConsumer_ConsumeHandlerErrorRequeues(t *testing.T) {
	acker := &stubAcker{}
	deliveries := make(chan amqp091.Delivery, 1)
	deliveries <- amqp091.Delivery{Body: []byte(`{"type":"order.created","payload":{}}`), Acknowledger: acker}
	close(deliveries)

	ch := &stubChannel{deliveries: deliveries}
	c := &rabbitMQConsumer{
		conn:    &stubConn{},
		channel: ch,
		queue:   "q",
		logger:  logger.NewNop(),
	}

	err := c.Consume(context.Background(), func(_ string, evt model.Event) error { return errors.New("temp") })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acker.acked {
		t.Fatal("did not expect ack on error")
	}
	if !acker.nacked || !acker.requeue {
		t.Fatal("expected nack with requeue on transient error")
	}
}

func TestRabbitMQConsumer_InvalidJSON(t *testing.T) {
	acker := &stubAcker{}
	deliveries := make(chan amqp091.Delivery, 1)
	deliveries <- amqp091.Delivery{Body: []byte(`{invalid`), Acknowledger: acker}
	close(deliveries)

	ch := &stubChannel{deliveries: deliveries}
	c := &rabbitMQConsumer{
		conn:    &stubConn{},
		channel: ch,
		queue:   "q",
		logger:  logger.NewNop(),
	}

	err := c.Consume(context.Background(), func(_ string, evt model.Event) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acker.requeue {
		t.Fatal("invalid JSON should not be requeued")
	}
}

func TestRabbitMQConsumer_ContextCancel(t *testing.T) {
	ch := &stubChannel{deliveries: make(chan amqp091.Delivery)}
	c := &rabbitMQConsumer{
		conn:    &stubConn{},
		channel: ch,
		queue:   "q",
		logger:  logger.NewNop(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	err := c.Consume(ctx, func(_ string, evt model.Event) error { return nil })
	if err == nil {
		t.Fatal("expected context error")
	}
	if time.Since(start) > time.Second {
		t.Fatal("consume did not stop promptly on cancel")
	}
}

func TestRabbitMQConsumer_PermanentErrorSendsToDLQ(t *testing.T) {
	acker := &stubAcker{}
	deliveries := make(chan amqp091.Delivery, 1)
	deliveries <- amqp091.Delivery{Body: []byte(`{"type":"order.created","payload":{}}`), Acknowledger: acker}
	close(deliveries)

	ch := &stubChannel{deliveries: deliveries}
	c := &rabbitMQConsumer{
		conn:    &stubConn{},
		channel: ch,
		queue:   "q",
		logger:  logger.NewNop(),
	}

	err := c.Consume(context.Background(), func(_ string, evt model.Event) error {
		return &client.PermanentError{Message: "nope", StatusCode: 404}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !acker.nacked || acker.requeue {
		t.Fatal("expected nack without requeue for permanent error")
	}
}

func TestRabbitMQConsumer_MaxRetriesSendsToDLQ(t *testing.T) {
	acker := &stubAcker{}
	deliveries := make(chan amqp091.Delivery, 1)
	deliveries <- amqp091.Delivery{
		Body:         []byte(`{"type":"order.created","payload":{}}`),
		Acknowledger: acker,
		Headers: amqp091.Table{
			"x-death": []interface{}{amqp091.Table{"count": int64(3)}},
		},
	}
	close(deliveries)

	ch := &stubChannel{deliveries: deliveries}
	c := &rabbitMQConsumer{
		conn:    &stubConn{},
		channel: ch,
		queue:   "q",
		logger:  logger.NewNop(),
	}

	err := c.Consume(context.Background(), func(_ string, evt model.Event) error { return errors.New("temporary") })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !acker.nacked || acker.requeue {
		t.Fatal("expected nack without requeue when max retries reached")
	}
}

func TestConsume_ExtractsEventIDFromHeader(t *testing.T) {
	acker := &stubAcker{}
	deliveries := make(chan amqp091.Delivery, 1)
	deliveries <- amqp091.Delivery{
		Body:         []byte(`{"type":"order.created","payload":{}}`),
		Acknowledger: acker,
		Headers:      amqp091.Table{"event_id": "evt-xyz"},
	}
	close(deliveries)

	ch := &stubChannel{deliveries: deliveries}
	c := &rabbitMQConsumer{
		conn:    &stubConn{},
		channel: ch,
		queue:   "q",
		logger:  logger.NewNop(),
	}

	var capturedEventID string
	if err := c.Consume(context.Background(), func(eventID string, evt model.Event) error {
		capturedEventID = eventID
		return nil
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedEventID != "evt-xyz" {
		t.Fatalf("expected event_id evt-xyz, got %q", capturedEventID)
	}
}

func TestExtractEventID_FallsBackToMessageIdThenHash(t *testing.T) {
	d1 := amqp091.Delivery{MessageId: "msg-1"}
	if id := extractEventID(d1); id != "msg-1" {
		t.Fatalf("got %q", id)
	}
	d2 := amqp091.Delivery{Body: []byte("hello")}
	if id := extractEventID(d2); id == "" {
		t.Fatal("expected fallback hash, got empty")
	}
	if extractEventID(d2) != extractEventID(amqp091.Delivery{Body: []byte("hello")}) {
		t.Fatal("hash should be deterministic")
	}
}

func TestRabbitMQConsumer_CloseClosesConnAndChannel(t *testing.T) {
	ch := &stubChannel{deliveries: make(chan amqp091.Delivery)}
	conn := &stubConn{}
	c := &rabbitMQConsumer{
		conn:    conn,
		channel: ch,
	}

	if err := c.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
	if !ch.closed {
		t.Fatal("expected channel to be closed")
	}
	if !conn.closed {
		t.Fatal("expected connection to be closed")
	}
}
