package model

import "encoding/json"

const (
	OrderCreated    string = "order.created"
	OrderProcessing string = "order.processing"
	OrderProcessed  string = "order.processed"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
