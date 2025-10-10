package model

import "encoding/json"

const (
	OrderCreated    = "order.created"
	OrderProcessing = "order.processing"
	OrderCompleted  = "order.completed"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
