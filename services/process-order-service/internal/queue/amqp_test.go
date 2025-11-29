package queue

import (
	"testing"

	"github.com/rabbitmq/amqp091-go"
)

type fakeRawConnection struct {
	closed bool
}

func (f *fakeRawConnection) Channel() (*amqp091.Channel, error) {
	return &amqp091.Channel{}, nil
}

func (f *fakeRawConnection) Close() error {
	f.closed = true
	return nil
}

func TestAMQPConnWrapperChannelAndClose(t *testing.T) {
	raw := &fakeRawConnection{}
	wrapper := &amqpConnWrapper{conn: raw}

	ch, err := wrapper.Channel()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ch == nil {
		t.Fatal("expected channel instance")
	}
	if err := wrapper.Close(); err != nil {
		t.Fatalf("close returned error: %v", err)
	}
	if !raw.closed {
		t.Fatal("expected raw connection close to be called")
	}
}
