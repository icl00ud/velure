package model

import "encoding/json"

const (
	OrderCreated    string = "order.created"
	OrderProcessing string = "order.processing"
	OrderCompleted  string = "order.completed"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
