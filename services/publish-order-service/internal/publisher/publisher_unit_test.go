package publisher

import (
	"testing"

	"github.com/icl00ud/publish-order-service/internal/model"
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
