package publisher

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/icl00ud/velure/services/publish-order-service/internal/model"
	"github.com/icl00ud/velure/shared/logger"
	"github.com/rabbitmq/amqp091-go"
)

// brokenChannel simulates a channel whose connection died after setup: every
// publish fails with the AMQP 504 "channel/connection is not open" error.
type brokenChannel struct {
	capturingChannel
}

func (c *brokenChannel) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
	return amqp091.ErrClosed
}

type brokenConn struct{ ch *brokenChannel }

func (c *brokenConn) Channel() (amqpPublisherChannel, error) { return c.ch, nil }
func (c *brokenConn) Close() error                           { return nil }

// A broker outage kills the publisher's channel; once the broker is back,
// PublishWithConfirm must redial and retry instead of failing forever
// (the outbox relay never restarts the process, so it would stay broken).
func TestPublishWithConfirm_ReconnectsOnClosedChannel(t *testing.T) {
	broken := &brokenChannel{}
	healthy := &capturingChannel{}
	dials := 0
	pub, err := newRabbitMQPublisher("amqp://test", "orders", logger.NewNop(),
		func(string) (amqpPublisherConn, error) {
			dials++
			if dials == 1 {
				return &brokenConn{ch: broken}, nil
			}
			return &capturingConn{ch: healthy}, nil
		})
	if err != nil {
		t.Fatalf("setup publisher failed: %v", err)
	}

	evt := model.OutboxEvent{
		ID:        "evt-1",
		EventType: "order.created",
		Payload:   json.RawMessage(`{"id":"order-1"}`),
	}

	if err := pub.PublishWithConfirm(context.Background(), evt); err != nil {
		t.Fatalf("publish after reconnect failed: %v", err)
	}
	if dials != 2 {
		t.Fatalf("dials = %d, want 2 (initial + reconnect)", dials)
	}
	if len(healthy.published) != 1 {
		t.Fatalf("published on healthy channel = %d, want 1", len(healthy.published))
	}
}
