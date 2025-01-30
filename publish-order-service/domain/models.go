package domain

import "time"

type CartItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type Order struct {
	ID        string     `json:"id"`
	Items     []CartItem `json:"items"`
	Total     int        `json:"total"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Product struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Price             int    `json:"price"`
	Disponibility     bool   `json:"disponibility"`
	QuantityWarehouse int    `json:"quantity_warehouse"`
}
