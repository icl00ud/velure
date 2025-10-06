// model/order.go
package model

import "time"

const (
	StatusCreated    = "CREATED"
	StatusProcessing = "PROCESSING"
	StatusCompleted  = "COMPLETED"
)

type Order struct {
	ID        string     `json:"id"`
	Items     []CartItem `json:"items"`
	Total     int        `json:"total"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
