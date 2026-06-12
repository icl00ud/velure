package queue

import (
	"context"
	"testing"
	"time"

	"github.com/icl00ud/velure/services/process-order-service/internal/model"
	"github.com/icl00ud/velure/shared/logger"
	"github.com/rabbitmq/amqp091-go"
)

// A broker restart closes the deliveries channel. Without a reconnect the
// worker returns silently and the service becomes a zombie: alive, consuming
// nothing, queue piling up. With a reconnectFn the consumer must redial and
// keep consuming.
func TestRabbitMQConsumer_ReconnectsWhenDeliveriesClose(t *testing.T) {
	dead := make(chan amqp091.Delivery)
	close(dead) // broker died: channel closes immediately

	acker := &stubAcker{}
	ctx, cancel := context.WithCancel(context.Background())

	fresh := make(chan amqp091.Delivery, 1)
	fresh <- amqp091.Delivery{
		Body:         []byte(`{"type":"order.created","payload":{}}`),
		Acknowledger: acker,
	}

	reconnects := 0
	c := &rabbitMQConsumer{
		conn:    &stubConn{},
		channel: &stubChannel{deliveries: dead},
		queue:   "q",
		logger:  logger.NewNop(),
		reconnectDelay: time.Millisecond,
		reconnect: func(ctx context.Context) (AMQPChannel, error) {
			reconnects++
			return &stubChannel{deliveries: fresh}, nil
		},
	}

	handled := 0
	err := c.Consume(ctx, func(_ context.Context, _ string, _ model.Event) error {
		handled++
		cancel() // stop after the post-reconnect message is processed
		return nil
	})
	if err != context.Canceled {
		t.Fatalf("Consume err = %v, want context.Canceled", err)
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
