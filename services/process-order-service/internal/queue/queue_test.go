package queue

import (
	"testing"
)

func TestRabbitMQConsumerCloseHandlesNil(t *testing.T) {
	c := &rabbitMQConsumer{}
	if err := c.Close(); err != nil {
		t.Fatalf("expected nil error on Close with nil fields, got %v", err)
	}
}

func TestRabbitPublisherCloseHandlesNil(t *testing.T) {
	p := &rabbitPublisher{}
	if err := p.Close(); err != nil {
		t.Fatalf("expected nil error on Close with nil fields, got %v", err)
	}
}
