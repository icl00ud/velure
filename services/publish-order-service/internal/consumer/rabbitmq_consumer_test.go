package consumer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/icl00ud/publish-order-service/internal/model"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type stubConsumerChannel struct {
	deliveries  chan amqp091.Delivery
	consumeErr  error
	qosCalled   bool
	closed      bool
	ackOnClose  bool
	declareErr  error
	bindErr     error
	queueName   string
	prefetchSet bool
}

func (s *stubConsumerChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error) {
	if s.consumeErr != nil {
		return nil, s.consumeErr
	}
	return s.deliveries, nil
}

func (s *stubConsumerChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp091.Table) error {
	return s.declareErr
}

func (s *stubConsumerChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp091.Table) (amqp091.Queue, error) {
	if s.declareErr != nil {
		return amqp091.Queue{}, s.declareErr
	}
	s.queueName = name
	return amqp091.Queue{Name: name}, nil
}

func (s *stubConsumerChannel) QueueBind(name, key, exchange string, noWait bool, args amqp091.Table) error {
	return s.bindErr
}

func (s *stubConsumerChannel) Qos(prefetchCount, prefetchSize int, global bool) error {
	s.prefetchSet = true
	return nil
}

func (s *stubConsumerChannel) Close() error {
	s.closed = true
	return nil
}

type stubConsumerConn struct {
	ch *stubConsumerChannel
}

func (s *stubConsumerConn) Channel() (*amqp091.Channel, error) {
	return nil, nil
}

func (s *stubConsumerConn) Close() error { return nil }

type stubAcker struct {
	acked   bool
	nacked  bool
	requeue bool
}

func (s *stubAcker) Ack(tag uint64, multiple bool) error {
	s.acked = true
	return nil
}

func (s *stubAcker) Nack(tag uint64, multiple bool, requeue bool) error {
	s.nacked = true
	s.requeue = requeue
	return nil
}

func (s *stubAcker) Reject(tag uint64, requeue bool) error {
	s.nacked = true
	s.requeue = requeue
	return nil
}

func TestRabbitConsumer_StartProcessesMessages(t *testing.T) {
	acker := &stubAcker{}
	msgs := make(chan amqp091.Delivery, 1)
	msgs <- amqp091.Delivery{Body: []byte(`{"type":"order.processing","payload":{}}`), Acknowledger: acker}
	close(msgs)

	ch := &stubConsumerChannel{deliveries: msgs}
	c := &rabbitConsumer{
		conn:    &stubConsumerConn{ch: ch},
		channel: ch,
		queue:   "q",
		handler: func(ctx context.Context, evt model.Event) error { return nil },
		logger:  zap.NewNop(),
		workers: 1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		_ = c.Start(ctx)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Start did not return after cancel")
	}

	if !acker.acked {
		t.Fatal("expected ack on success")
	}
}

func TestRabbitConsumer_HandlerErrorNacks(t *testing.T) {
	acker := &stubAcker{}
	msgs := make(chan amqp091.Delivery, 1)
	msgs <- amqp091.Delivery{Body: []byte(`{"type":"order.processing","payload":{}}`), Acknowledger: acker}
	close(msgs)

	ch := &stubConsumerChannel{deliveries: msgs}
	c := &rabbitConsumer{
		conn:    &stubConsumerConn{ch: ch},
		channel: ch,
		queue:   "q",
		handler: func(ctx context.Context, evt model.Event) error { return errors.New("fail") },
		logger:  zap.NewNop(),
		workers: 1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		_ = c.Start(ctx)
		close(done)
	}()
	time.Sleep(20 * time.Millisecond)
	cancel()
	<-done

	if !acker.nacked || !acker.requeue {
		t.Fatal("expected nack with requeue on handler error")
	}
}

func TestRabbitConsumer_ProcessMessageInvalidJSON(t *testing.T) {
	acker := &stubAcker{}
	ch := &stubConsumerChannel{}
	c := &rabbitConsumer{
		channel: ch,
		logger:  zap.NewNop(),
	}
	msg := amqp091.Delivery{Body: []byte(`{invalid`), Acknowledger: acker}
	if err := c.processMessage(context.Background(), msg); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if acker.requeue {
		t.Fatal("invalid JSON should not requeue")
	}
}

func TestRabbitConsumer_CloseClosesChannelAndConn(t *testing.T) {
	ch := &stubConsumerChannel{deliveries: make(chan amqp091.Delivery)}
	conn := &stubConsumerConn{ch: ch}
	c := &rabbitConsumer{
		conn:    conn,
		channel: ch,
		queue:   "q",
		logger:  zap.NewNop(),
	}

	if err := c.Close(); err != nil {
		t.Fatalf("close returned error: %v", err)
	}
	if !ch.closed {
		t.Fatal("expected channel closed")
	}
}
