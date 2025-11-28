package publisher

import (
	"errors"
	"sync"
	"testing"

	"github.com/icl00ud/publish-order-service/internal/model"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

func TestPublish_ReturnsErrorWhenClosed(t *testing.T) {
	pub := &rabbitMQPublisher{
		logger: zap.NewNop(),
		closed: true,
	}

	err := pub.Publish(model.Event{Type: "order.test", Payload: []byte(`{}`)})
	if err == nil {
		t.Fatal("expected error when publishing after Close")
	}
}

func TestClose_IsIdempotentWithoutConnections(t *testing.T) {
	pub := &rabbitMQPublisher{
		logger: zap.NewNop(),
	}

	if err := pub.Close(); err != nil {
		t.Fatalf("first close returned error: %v", err)
	}
	if err := pub.Close(); err != nil {
		t.Fatalf("second close returned error: %v", err)
	}
}

type fakePubChannel struct {
	publishErr error
	published  int
	closed     bool
}

func (f *fakePubChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
	f.published++
	return f.publishErr
}
func (f *fakePubChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp091.Table) error {
	return nil
}
func (f *fakePubChannel) Close() error {
	f.closed = true
	return nil
}

type fakePubConn struct {
	ch     *fakePubChannel
	closed bool
}

func (f *fakePubConn) Channel() (*amqp091.Channel, error) { return nil, nil }
func (f *fakePubConn) Close() error {
	f.closed = true
	return nil
}

func TestPublish_ReconnectsOnError(t *testing.T) {
	first := &fakePubChannel{publishErr: errors.New("fail")}
	second := &fakePubChannel{}
	pub := &rabbitMQPublisher{}
	pub.logger = zap.NewNop()
	pub.exchange = "ex"
	pub.ch = first
	pub.connectFn = func() error {
		pub.ch = second
		return nil
	}

	err := pub.Publish(model.Event{Type: "order.created", Payload: []byte(`{}`)})
	if err != nil {
		t.Fatalf("expected publish to succeed after reconnect, got %v", err)
	}
	if first.published != 1 {
		t.Fatalf("expected first channel to be used once, got %d", first.published)
	}
	if second.published != 1 {
		t.Fatalf("expected second channel to publish once, got %d", second.published)
	}
}

func TestPublish_ReconnectFailure(t *testing.T) {
	ch := &fakePubChannel{publishErr: errors.New("fail")}
	pub := &rabbitMQPublisher{
		logger:   zap.NewNop(),
		exchange: "ex",
		ch:       ch,
		connectFn: func() error {
			return errors.New("reconnect fail")
		},
	}

	err := pub.Publish(model.Event{Type: "order.created", Payload: []byte(`{}`)})
	if err == nil {
		t.Fatal("expected error when reconnect fails")
	}
}

func TestPublish_ReconnectsWhenChannelNil(t *testing.T) {
	// Build publisher with nil channel to force reconnect path
	pub := &rabbitMQPublisher{
		logger:  zap.NewNop(),
		amqpURL: "amqp://invalid", // will make connect() fail
		mu:      sync.Mutex{},
		connectFn: func() error {
			return errors.New("dial fail")
		},
	}

	pub.ch = nil
	pub.conn = nil

	err := pub.Publish(model.Event{Type: "order.test", Payload: []byte(`{}`)})
	if err == nil {
		t.Fatal("expected error when reconnect fails")
	}
}
