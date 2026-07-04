package amqp

import (
	"testing"

	"github.com/rabbitmq/amqp091-go"
)

func TestHeadersToMap(t *testing.T) {
	headers := amqp091.Table{
		"traceparent": "00-abc-def-01",
		"non-string":  int64(42),
	}
	got := HeadersToMap(headers)
	if got["traceparent"] != "00-abc-def-01" {
		t.Fatalf("expected traceparent to be copied, got %v", got)
	}
	if _, ok := got["non-string"]; ok {
		t.Fatalf("expected non-string header to be dropped, got %v", got)
	}
}

func TestRetryCount(t *testing.T) {
	cases := []struct {
		name    string
		headers amqp091.Table
		want    int64
	}{
		{"nil headers", nil, 0},
		{"no x-death", amqp091.Table{}, 0},
		{"malformed x-death", amqp091.Table{"x-death": "not-a-slice"}, 0},
		{
			"valid x-death",
			amqp091.Table{"x-death": []any{amqp091.Table{"count": int64(2)}}},
			2,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := RetryCount(tc.headers); got != tc.want {
				t.Fatalf("RetryCount() = %d, want %d", got, tc.want)
			}
		})
	}
}
