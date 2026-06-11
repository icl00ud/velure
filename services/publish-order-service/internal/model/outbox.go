package model

import (
	"encoding/json"
	"time"
)

// OutboxEvent is a domain event awaiting publication to RabbitMQ.
// PublishedAt nil = pending; non-nil = already published (kept for audit).
type OutboxEvent struct {
	ID          string          `json:"id"`
	AggregateID string          `json:"aggregate_id"`
	EventType   string          `json:"event_type"`
	Payload     json.RawMessage `json:"payload"`
	CreatedAt   time.Time       `json:"created_at"`
	PublishedAt *time.Time      `json:"published_at,omitempty"`
	// TraceContext carries the W3C traceparent of the request that created the
	// event, so the relay can continue the trace across the outbox boundary.
	TraceContext string `json:"trace_context,omitempty"`
}
