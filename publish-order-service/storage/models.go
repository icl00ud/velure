package storage

import (
	"time"

	"github.com/icl00ud/publish-order-service/domain"
)

// Order representa um pedido no banco de dados.
type Order struct {
	ID        string            `json:"id"`
	Items     []domain.CartItem `json:"items"`
	Total     int               `json:"total"`
	Status    string            `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}
