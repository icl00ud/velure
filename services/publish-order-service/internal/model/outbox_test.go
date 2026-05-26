package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestOutboxEvent_JSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	evt := OutboxEvent{
		ID:          "evt-1",
		AggregateID: "order-1",
		EventType:   OrderCreated,
		Payload:     json.RawMessage(`{"id":"order-1"}`),
		CreatedAt:   now,
	}
	b, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got OutboxEvent
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ID != evt.ID || got.AggregateID != evt.AggregateID || got.EventType != evt.EventType {
		t.Fatalf("mismatch: %+v", got)
	}
	if !got.CreatedAt.Equal(evt.CreatedAt) {
		t.Fatalf("created_at mismatch: got %v want %v", got.CreatedAt, evt.CreatedAt)
	}
	if got.PublishedAt != nil {
		t.Fatalf("expected nil PublishedAt, got %v", got.PublishedAt)
	}
}
