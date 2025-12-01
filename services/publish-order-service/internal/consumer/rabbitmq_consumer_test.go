package consumer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/icl00ud/publish-order-service/internal/model"
	"github.com/rabbitmq/amqp091-go"
	"github.com/icl00ud/velure-shared/logger"
)

type stubConsumerChannel struct {
	deliveries      chan amqp091.Delivery
	consumeErr      error
	qosCalled       bool
	closed          bool
	ackOnClose      bool
	declareErr      error
	queueDeclareErr error
	bindErr         error
	queueName       string
	prefetchSet     bool
	declaredEx      bool
	bindKeys        []string
	qosErr          error
}

func (s *stubConsumerChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error) {
	if s.consumeErr != nil {
		return nil, s.consumeErr
	}
	return s.deliveries, nil
}

func (s *stubConsumerChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp091.Table) error {
	s.declaredEx = true
	return s.declareErr
}

func (s *stubConsumerChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp091.Table) (amqp091.Queue, error) {
	if s.queueDeclareErr != nil {
		return amqp091.Queue{}, s.queueDeclareErr
	}
	s.queueName = name
	return amqp091.Queue{Name: name}, nil
}

func (s *stubConsumerChannel) QueueBind(name, key, exchange string, noWait bool, args amqp091.Table) error {
	s.bindKeys = append(s.bindKeys, key)
	return s.bindErr
}

func (s *stubConsumerChannel) Qos(prefetchCount, prefetchSize int, global bool) error {
	if s.qosErr != nil {
		return s.qosErr
	}
	s.prefetchSet = true
	return nil
}

func (s *stubConsumerChannel) Close() error {
	s.closed = true
	return nil
}

type stubConsumerConn struct {
	ch         amqpChan
	closed     bool
	channelErr error
}

func (s *stubConsumerConn) Channel() (amqpChan, error) {
	if s.channelErr != nil {
		return nil, s.channelErr
	}
	return s.ch, nil
}

func (s *stubConsumerConn) Close() error { s.closed = true; return nil }

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

func TestNewRabbitMQConsumer_InitializesBindings(t *testing.T) {
	origDial := dialRabbitMQ
	defer func() { dialRabbitMQ = origDial }()

	ch := &stubConsumerChannel{}
	conn := &stubConsumerConn{ch: ch}
	dialRabbitMQ = func(amqpURL string) (amqpConn, error) {
		return conn, nil
	}

	cons, err := NewRabbitMQConsumer(
		"amqp://example",
		"orders.exchange",
		"orders.queue",
		func(ctx context.Context, evt model.Event) error { return nil },
		2,
		logger.NewNop(),
	)
	if err != nil {
		t.Fatalf("expected constructor to succeed, got %v", err)
	}
	if cons == nil {
		t.Fatal("expected non-nil consumer")
	}
	if !ch.declaredEx {
		t.Fatal("expected exchange declaration")
	}
	if ch.queueName != "orders.queue" {
		t.Fatalf("expected queue name set, got %s", ch.queueName)
	}
	if len(ch.bindKeys) != 2 || ch.bindKeys[0] != "order.processing" || ch.bindKeys[1] != "order.completed" {
		t.Fatalf("queue bindings not applied: %v", ch.bindKeys)
	}
	if !ch.prefetchSet {
		t.Fatal("expected QoS to be configured")
	}
}

func TestNewRabbitMQConsumer_ChannelErrorClosesConn(t *testing.T) {
	origDial := dialRabbitMQ
	defer func() { dialRabbitMQ = origDial }()

	conn := &stubConsumerConn{channelErr: errors.New("channel open failed")}
	dialRabbitMQ = func(amqpURL string) (amqpConn, error) {
		return conn, nil
	}

	if _, err := NewRabbitMQConsumer("amqp://example", "ex", "q", func(context.Context, model.Event) error { return nil }, 1, logger.NewNop()); err == nil {
		t.Fatal("expected error when channel creation fails")
	}
	if !conn.closed {
		t.Fatal("expected connection to close on channel error")
	}
}

func TestNewRabbitMQConsumer_ExchangeDeclareError(t *testing.T) {
	origDial := dialRabbitMQ
	defer func() { dialRabbitMQ = origDial }()

	ch := &stubConsumerChannel{declareErr: errors.New("declare fail")}
	conn := &stubConsumerConn{ch: ch}
	dialRabbitMQ = func(amqpURL string) (amqpConn, error) {
		return conn, nil
	}

	if _, err := NewRabbitMQConsumer("amqp://example", "ex", "q", func(context.Context, model.Event) error { return nil }, 1, logger.NewNop()); err == nil {
		t.Fatal("expected error when exchange declaration fails")
	}
	if !ch.closed || !conn.closed {
		t.Fatal("expected resources to close on setup failure")
	}
}

func TestNewRabbitMQConsumer_QueueDeclareError(t *testing.T) {
	origDial := dialRabbitMQ
	defer func() { dialRabbitMQ = origDial }()

	ch := &stubConsumerChannel{queueDeclareErr: errors.New("queue fail")}
	conn := &stubConsumerConn{ch: ch}
	dialRabbitMQ = func(amqpURL string) (amqpConn, error) {
		return conn, nil
	}

	if _, err := NewRabbitMQConsumer("amqp://example", "ex", "q", func(context.Context, model.Event) error { return nil }, 1, logger.NewNop()); err == nil {
		t.Fatal("expected error when queue declaration fails")
	}
	if !ch.closed || !conn.closed {
		t.Fatal("expected resources to close on setup failure")
	}
}

func TestNewRabbitMQConsumer_QueueBindError(t *testing.T) {
	origDial := dialRabbitMQ
	defer func() { dialRabbitMQ = origDial }()

	ch := &stubConsumerChannel{bindErr: errors.New("bind fail")}
	conn := &stubConsumerConn{ch: ch}
	dialRabbitMQ = func(amqpURL string) (amqpConn, error) {
		return conn, nil
	}

	if _, err := NewRabbitMQConsumer("amqp://example", "ex", "q", func(context.Context, model.Event) error { return nil }, 1, logger.NewNop()); err == nil {
		t.Fatal("expected error when queue binding fails")
	}
	if !ch.closed || !conn.closed {
		t.Fatal("expected resources to close on binding failure")
	}
}

func TestNewRabbitMQConsumer_QosError(t *testing.T) {
	origDial := dialRabbitMQ
	defer func() { dialRabbitMQ = origDial }()

	ch := &stubConsumerChannel{qosErr: errors.New("qos fail")}
	conn := &stubConsumerConn{ch: ch}
	dialRabbitMQ = func(amqpURL string) (amqpConn, error) {
		return conn, nil
	}

	if _, err := NewRabbitMQConsumer("amqp://example", "ex", "q", func(context.Context, model.Event) error { return nil }, 1, logger.NewNop()); err == nil {
		t.Fatal("expected error when QoS setup fails")
	}
	if !ch.closed || !conn.closed {
		t.Fatal("expected resources to close on QoS failure")
	}
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
		logger:  logger.NewNop(),
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
		logger:  logger.NewNop(),
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
		logger:  logger.NewNop(),
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
		logger:  logger.NewNop(),
	}

	if err := c.Close(); err != nil {
		t.Fatalf("close returned error: %v", err)
	}
	if !ch.closed {
		t.Fatal("expected channel closed")
	}
}

func TestLiveConsumerConn_AllowsNil(t *testing.T) {
	conn := &liveConsumerConn{}
	if _, err := conn.Channel(); err == nil {
		t.Fatal("expected error when channel requested on nil connection")
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("expected nil error closing nil connection, got %v", err)
	}
}
