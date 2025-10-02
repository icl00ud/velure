package models

// PaginatedResponse representa uma resposta paginada genérica
// T é o tipo dos items retornados
type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"`
	TotalCount int64 `json:"totalCount"`
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalPages int   `json:"totalPages"`
}

// NewPaginatedResponse cria uma nova resposta paginada
func NewPaginatedResponse[T any](data []T, totalCount int64, page, pageSize int) *PaginatedResponse[T] {
	totalPages := int(totalCount) / pageSize
	if int(totalCount)%pageSize > 0 {
		totalPages++
	}

	return &PaginatedResponse[T]{
		Data:       data,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
