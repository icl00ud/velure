package queue

import (
	"errors"
	"testing"

	"github.com/icl00ud/process-order-service/internal/model"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type fakeChannel struct {
	published bool
	closed    bool
	err       error
	exchange  string
	key       string
	body      []byte
}

func (f *fakeChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error) {
	return nil, nil
}

func (f *fakeChannel) Qos(prefetchCount, prefetchSize int, global bool) error {
	return nil
}

func (f *fakeChannel) QueueBind(queue, key, exchange string, noWait bool, args amqp091.Table) error {
	return nil
}

func (f *fakeChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
	f.published = true
	f.exchange = exchange
	f.key = key
	f.body = msg.Body
	return f.err
}

func (f *fakeChannel) Close() error {
	f.closed = true
	return nil
}

func (f *fakeChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp091.Table) error {
	return nil
}

type fakeConn struct {
	ch     *fakeChannel
	closed bool
}

func (f *fakeConn) Channel() (AMQPChannel, error) {
	// Not used in tests because we inject channel directly.
	return nil, nil
}

func (f *fakeConn) Close() error {
	f.closed = true
	return nil
}

func TestRabbitPublisher_PublishSuccess(t *testing.T) {
	ch := &fakeChannel{}
	pub := &rabbitPublisher{
		channel:  ch,
		exchange: "orders",
		logger:   zap.NewNop(),
	}

	evt := model.Event{Type: "order.created", Payload: []byte(`{"id":"1"}`)}
	if err := pub.Publish(evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ch.published {
		t.Fatal("expected publish to be called")
	}
	if ch.exchange != "orders" || ch.key != evt.Type {
		t.Fatalf("unexpected routing: %s %s", ch.exchange, ch.key)
	}
	if string(ch.body) == "" {
		t.Fatal("expected body to be marshaled")
	}
}

func TestRabbitPublisher_PublishError(t *testing.T) {
	ch := &fakeChannel{err: errors.New("publish failed")}
	pub := &rabbitPublisher{
		channel:  ch,
		exchange: "orders",
		logger:   zap.NewNop(),
	}

	err := pub.Publish(model.Event{Type: "order.created", Payload: []byte(`{}`)})
	if err == nil {
		t.Fatal("expected error from publish")
	}
}

func TestRabbitPublisher_CloseClosesResources(t *testing.T) {
	ch := &fakeChannel{}
	conn := &fakeConn{}
	pub := &rabbitPublisher{
		channel: ch,
		conn:    conn,
		logger:  zap.NewNop(),
	}

	if err := pub.Close(); err != nil {
		t.Fatalf("close returned error: %v", err)
	}
	if !ch.closed {
		t.Fatal("expected channel to be closed")
	}
	if !conn.closed {
		t.Fatal("expected connection to be closed")
	}

	// Ensure idempotency
	if err := pub.Close(); err != nil {
		t.Fatalf("second close returned error: %v", err)
	}
}
