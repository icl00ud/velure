package domain

type EventType string

const (
	OrderCreated   EventType = "order.created"
	OrderUpdated   EventType = "order.updated"
	OrderCancelled EventType = "order.cancelled"
)

type Event struct {
	Type  EventType `json:"type"`
	Order Order     `json:"order"`
}
