package domain

import "time"

type PaymentStatus string

const (
	PaymentProcessed PaymentStatus = "processed"
	PaymentFailed    PaymentStatus = "failed"
)

type Payment struct {
	ID          string        `json:"id"`
	OrderID     string        `json:"order_id"`
	Amount      int           `json:"amount"`
	Status      PaymentStatus `json:"status"`
	ProcessedAt time.Time     `json:"processed_at"`
}
