package queue

import (
	"testing"

	"github.com/rabbitmq/amqp091-go"
)

func TestGetRetryCount(t *testing.T) {
	tests := []struct {
		name    string
		headers amqp091.Table
		want    int64
	}{
		{
			name:    "nil headers",
			headers: nil,
			want:    0,
		},
		{
			name:    "missing x-death",
			headers: amqp091.Table{"foo": "bar"},
			want:    0,
		},
		{
			name:    "x-death empty slice",
			headers: amqp091.Table{"x-death": []interface{}{}},
			want:    0,
		},
		{
			name: "x-death without count",
			headers: amqp091.Table{
				"x-death": []interface{}{amqp091.Table{"reason": "rejected"}},
			},
			want: 0,
		},
		{
			name: "valid retry count",
			headers: amqp091.Table{
				"x-death": []interface{}{
					amqp091.Table{"count": int64(2)},
				},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRetryCount(tt.headers); got != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

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
