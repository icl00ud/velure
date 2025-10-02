package models

// PaginatedUsersResponse representa uma resposta paginada de usuários
type PaginatedUsersResponse struct {
	Users      []UserResponse `json:"users"`
	TotalCount int64          `json:"totalCount"`
	Page       int            `json:"page"`
	PageSize   int            `json:"pageSize"`
	TotalPages int            `json:"totalPages"`
}

// NewPaginatedUsersResponse cria uma nova resposta paginada de usuários
func NewPaginatedUsersResponse(users []UserResponse, totalCount int64, page, pageSize int) *PaginatedUsersResponse {
	totalPages := int(totalCount) / pageSize
	if int(totalCount)%pageSize > 0 {
		totalPages++
	}

	return &PaginatedUsersResponse{
		Users:      users,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
