package queue

import (
	"context"
	"testing"

	"github.com/icl00ud/velure/services/process-order-service/internal/model"
	"github.com/icl00ud/velure/shared/logger"
	"github.com/rabbitmq/amqp091-go"
)

type failingPublishChannel struct {
	stubChannel
	calls int
}

func (f *failingPublishChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
	f.calls++
	return amqp091.ErrClosed
}

type countingPublishChannel struct {
	stubChannel
	published []amqp091.Publishing
}

func (c *countingPublishChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
	c.published = append(c.published, msg)
	return nil
}

// A broker restart kills the publisher channel; status events (order.failed,
// order.completed) were silently lost and orders stayed CREATED forever.
// Publish must redial and retry once.
func TestRabbitPublisher_ReconnectsOnClosedChannel(t *testing.T) {
	dead := &failingPublishChannel{}
	healthy := &countingPublishChannel{}

	reconnects := 0
	p := &rabbitPublisher{
		channel:  dead,
		exchange: "orders",
		logger:   logger.NewNop(),
		reconnect: func() (AMQPChannel, error) {
			reconnects++
			return healthy, nil
		},
	}

	evt := model.Event{Type: "order.failed", Payload: []byte(`{}`)}
	if err := p.Publish(context.Background(), evt); err != nil {
		t.Fatalf("publish after reconnect failed: %v", err)
	}
	if reconnects != 1 {
		t.Fatalf("reconnects = %d, want 1", reconnects)
	}
	if len(healthy.published) != 1 {
		t.Fatalf("published on healthy channel = %d, want 1", len(healthy.published))
	}
}
