package domain

type Order struct {
	ID          string `json:"id"`
	ProductID   string `json:"product_id"`
	Quantity    int    `json:"quantity"`
	TotalAmount int    `json:"total_amount"`
}
