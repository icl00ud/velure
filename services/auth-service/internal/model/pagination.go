package models

// PaginatedUsersResponse represents a paginated users response
type PaginatedUsersResponse struct {
	Users      []UserResponse `json:"users"`
	TotalCount int64          `json:"totalCount"`
	Page       int            `json:"page"`
	PageSize   int            `json:"pageSize"`
	TotalPages int            `json:"totalPages"`
}

// NewPaginatedUsersResponse builds a new paginated users response
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
