package consumer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/icl00ud/publish-order-service/internal/model"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type stubAcknowledger struct {
	acked        bool
	nacked       bool
	requeueFlag  bool
	tag          uint64
	multipleAck  bool
	multipleNack bool
}

func (s *stubAcknowledger) Ack(tag uint64, multiple bool) error {
	s.acked = true
	s.tag = tag
	s.multipleAck = multiple
	return nil
}

func (s *stubAcknowledger) Nack(tag uint64, multiple bool, requeue bool) error {
	s.nacked = true
	s.requeueFlag = requeue
	s.tag = tag
	s.multipleNack = multiple
	return nil
}

func (s *stubAcknowledger) Reject(tag uint64, requeue bool) error {
	s.nacked = true
	s.requeueFlag = requeue
	s.tag = tag
	return nil
}

func TestProcessMessage_Success(t *testing.T) {
	t.Helper()
	var handled bool
	rc := &rabbitConsumer{
		handler: func(ctx context.Context, evt model.Event) error {
			handled = true
			if evt.Type != model.OrderCompleted {
				t.Fatalf("unexpected event type %s", evt.Type)
			}
			if len(evt.Payload) == 0 {
				t.Fatal("payload should not be empty")
			}
			return nil
		},
		logger: zap.NewNop(),
	}

	msg := amqp091.Delivery{Body: []byte(`{"type":"order.completed","payload":{"id":"123"}}`)}

	if err := rc.processMessage(context.Background(), msg); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if !handled {
		t.Fatal("handler was not invoked")
	}
}

func TestProcessMessage_InvalidJSON(t *testing.T) {
	rc := &rabbitConsumer{
		handler: func(ctx context.Context, evt model.Event) error {
			t.Fatal("handler should not be called for invalid JSON")
			return nil
		},
		logger: zap.NewNop(),
	}

	msg := amqp091.Delivery{Body: []byte(`{"type":`)} // malformed JSON

	if err := rc.processMessage(context.Background(), msg); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestProcessMessage_HandlerError(t *testing.T) {
	expected := errors.New("boom")
	rc := &rabbitConsumer{
		handler: func(ctx context.Context, evt model.Event) error {
			return expected
		},
		logger: zap.NewNop(),
	}

	msg := amqp091.Delivery{Body: []byte(`{"type":"order.processing","payload":{}}`)}

	if err := rc.processMessage(context.Background(), msg); !errors.Is(err, expected) {
		t.Fatalf("expected handler error, got %v", err)
	}
}

func TestWorker_AckOnSuccess(t *testing.T) {
	ack := &stubAcknowledger{}
	rc := &rabbitConsumer{
		handler: func(ctx context.Context, evt model.Event) error { return nil },
		logger:  zap.NewNop(),
	}

	msgs := make(chan amqp091.Delivery, 1)
	msgs <- amqp091.Delivery{
		Body:         []byte(`{"type":"order.processing","payload":{}}`),
		Acknowledger: ack,
		DeliveryTag:  7,
	}
	close(msgs)

	rc.worker(context.Background(), 1, msgs)

	if !ack.acked {
		t.Fatal("expected message to be acknowledged")
	}
	if ack.tag != 7 {
		t.Fatalf("expected ack tag 7, got %d", ack.tag)
	}
}

func TestWorker_NackOnError(t *testing.T) {
	ack := &stubAcknowledger{}
	rc := &rabbitConsumer{
		handler: func(ctx context.Context, evt model.Event) error { return errors.New("fail") },
		logger:  zap.NewNop(),
	}

	msgs := make(chan amqp091.Delivery, 1)
	msgs <- amqp091.Delivery{
		Body:         []byte(`{"type":"order.processing","payload":{}}`),
		Acknowledger: ack,
		DeliveryTag:  9,
	}
	close(msgs)

	rc.worker(context.Background(), 2, msgs)

	if !ack.nacked {
		t.Fatal("expected message to be negatively acknowledged")
	}
	if !ack.requeueFlag {
		t.Fatal("expected message to be requeued on failure")
	}
	if ack.tag != 9 {
		t.Fatalf("expected nack tag 9, got %d", ack.tag)
	}
	if ack.multipleNack {
		t.Fatal("did not expect multiple nack")
	}
}

func TestWorker_ReturnsWhenChannelClosed(t *testing.T) {
	rc := &rabbitConsumer{
		handler: func(ctx context.Context, evt model.Event) error {
			t.Fatal("handler should not be called")
			return nil
		},
		logger: zap.NewNop(),
	}

	msgs := make(chan amqp091.Delivery)
	close(msgs)

	done := make(chan struct{})
	go func() {
		rc.worker(context.Background(), 3, msgs)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not return after channel closed")
	}
}

func TestWorker_StopsOnContextCancel(t *testing.T) {
	rc := &rabbitConsumer{
		handler: func(ctx context.Context, evt model.Event) error { return nil },
		logger:  zap.NewNop(),
	}

	msgs := make(chan amqp091.Delivery)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		rc.worker(ctx, 4, msgs)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after context cancel")
	}
}
