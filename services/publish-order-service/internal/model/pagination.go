// model/pagination.go
package model

// PaginatedResponse representa uma resposta paginada genérica
type PaginatedOrdersResponse struct {
	Orders     []Order `json:"orders"`
	TotalCount int64   `json:"totalCount"`
	Page       int     `json:"page"`
	PageSize   int     `json:"pageSize"`
	TotalPages int     `json:"totalPages"`
}

// NewPaginatedOrdersResponse cria uma nova resposta paginada de orders
func NewPaginatedOrdersResponse(orders []Order, totalCount int64, page, pageSize int) *PaginatedOrdersResponse {
	totalPages := int(totalCount) / pageSize
	if int(totalCount)%pageSize > 0 {
		totalPages++
	}

	return &PaginatedOrdersResponse{
		Orders:     orders,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
