package queue

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/icl00ud/process-order-service/internal/model"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type stubChannel struct {
	deliveries chan amqp091.Delivery
	consumeErr error
	qosCalled  bool
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

func (s *stubChannel) Close() error {
	close(s.deliveries)
	return nil
}

type stubConn struct{}

func (s *stubConn) Channel() (*amqp091.Channel, error) { return nil, nil }
func (s *stubConn) Close() error                       { return nil }

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
		logger:  zap.NewNop(),
	}

	if err := c.Consume(context.Background(), func(evt model.Event) error { return nil }); err != nil {
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
		logger:  zap.NewNop(),
	}

	err := c.Consume(context.Background(), func(evt model.Event) error { return errors.New("temp") })
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
		logger:  zap.NewNop(),
	}

	err := c.Consume(context.Background(), func(evt model.Event) error { return nil })
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
		logger:  zap.NewNop(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	err := c.Consume(ctx, func(evt model.Event) error { return nil })
	if err == nil {
		t.Fatal("expected context error")
	}
	if time.Since(start) > time.Second {
		t.Fatal("consume did not stop promptly on cancel")
	}
}
