package consumer

import (
	"context"
	"testing"
	"time"

	"github.com/icl00ud/velure/services/publish-order-service/internal/model"
	"github.com/icl00ud/velure/shared/logger"
	"github.com/rabbitmq/amqp091-go"
)

// A broker restart closes the deliveries channel; workers logged "message
// channel closed" and returned, leaving Start blocked on ctx forever — a
// zombie that never consumes status updates again. Start must redial and
// restart the workers.
func TestRabbitConsumer_Start_ReconnectsWhenDeliveriesClose(t *testing.T) {
	dead := make(chan amqp091.Delivery)
	close(dead)

	acker := &stubAcker{}
	fresh := make(chan amqp091.Delivery, 1)
	fresh <- amqp091.Delivery{
		Body:         []byte(`{"type":"order.completed","payload":{}}`),
		Acknowledger: acker,
	}

	ctx, cancel := context.WithCancel(context.Background())
	handled := 0

	reconnects := 0
	c := &rabbitConsumer{
		conn:    &stubConsumerConn{},
		channel: &stubConsumerChannel{deliveries: dead},
		queue:   "q",
		logger:  logger.NewNop(),
		workers: 2,
		handler: func(_ context.Context, _ model.Event) error {
			handled++
			cancel()
			return nil
		},
		reconnectDelay: time.Millisecond,
		reconnect: func(ctx context.Context) (amqpChan, error) {
			reconnects++
			return &stubConsumerChannel{deliveries: fresh}, nil
		},
	}

	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if reconnects != 1 {
		t.Fatalf("reconnects = %d, want 1", reconnects)
	}
	if handled != 1 {
		t.Fatalf("handled = %d, want 1", handled)
	}
	if !acker.acked {
		t.Fatal("expected ack on post-reconnect message")
	}
}
