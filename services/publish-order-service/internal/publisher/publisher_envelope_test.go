package publisher

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/icl00ud/velure/services/publish-order-service/internal/model"
	"github.com/icl00ud/velure/shared/logger"
	"github.com/rabbitmq/amqp091-go"
)

type capturingChannel struct {
	published []amqp091.Publishing
	confirms  chan amqp091.Confirmation
}

func (c *capturingChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
	c.published = append(c.published, msg)
	return nil
}

func (c *capturingChannel) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
	c.published = append(c.published, msg)
	go func() { c.confirms <- amqp091.Confirmation{DeliveryTag: 1, Ack: true} }()
	return nil
}

func (c *capturingChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp091.Table) error {
	return nil
}

func (c *capturingChannel) Confirm(noWait bool) error { return nil }

func (c *capturingChannel) NotifyPublish(confirm chan amqp091.Confirmation) chan amqp091.Confirmation {
	c.confirms = confirm
	return confirm
}

func (c *capturingChannel) Close() error { return nil }

type capturingConn struct{ ch *capturingChannel }

func (c *capturingConn) Channel() (amqpPublisherChannel, error) { return c.ch, nil }
func (c *capturingConn) Close() error                           { return nil }

// The outbox stores the bare aggregate payload; consumers expect the
// {"type","payload"} envelope. PublishWithConfirm must wrap the payload,
// otherwise consumers see an empty event type and silently drop the message.
func TestPublishWithConfirm_WrapsPayloadInEventEnvelope(t *testing.T) {
	ch := &capturingChannel{}
	pub, err := newRabbitMQPublisher("amqp://test", "orders", logger.NewNop(),
		func(string) (amqpPublisherConn, error) { return &capturingConn{ch: ch}, nil })
	if err != nil {
		t.Fatalf("setup publisher failed: %v", err)
	}

	evt := model.OutboxEvent{
		ID:        "evt-1",
		EventType: "order.created",
		Payload:   json.RawMessage(`{"id":"order-1","total":10.5}`),
		CreatedAt: time.Now(),
	}

	if err := pub.PublishWithConfirm(context.Background(), evt); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	if len(ch.published) != 1 {
		t.Fatalf("expected 1 published message, got %d", len(ch.published))
	}

	var got model.Event
	if err := json.Unmarshal(ch.published[0].Body, &got); err != nil {
		t.Fatalf("body is not an event envelope: %v", err)
	}
	if got.Type != "order.created" {
		t.Errorf("envelope type: want order.created, got %q", got.Type)
	}

	var payload struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(got.Payload, &payload); err != nil {
		t.Fatalf("payload unmarshal failed: %v", err)
	}
	if payload.ID != "order-1" {
		t.Errorf("payload id: want order-1, got %q", payload.ID)
	}
}
