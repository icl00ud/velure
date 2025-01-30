package domain

import "encoding/json"

type EventType string

const (
	OrderCreated EventType = "order.created"
)

type Event struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
