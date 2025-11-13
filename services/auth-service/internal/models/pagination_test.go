package models

import (
	"testing"
	"time"
)

func TestNewPaginatedUsersResponse(t *testing.T) {
	now := time.Now()
	users := []UserResponse{
		{ID: 1, Name: "User 1", Email: "user1@example.com", CreatedAt: now, UpdatedAt: now},
		{ID: 2, Name: "User 2", Email: "user2@example.com", CreatedAt: now, UpdatedAt: now},
		{ID: 3, Name: "User 3", Email: "user3@example.com", CreatedAt: now, UpdatedAt: now},
	}

	tests := []struct {
		name           string
		users          []UserResponse
		totalCount     int64
		page           int
		pageSize       int
		expectedPages  int
	}{
		{
			name:          "3 users, page 1, pageSize 10",
			users:         users,
			totalCount:    3,
			page:          1,
			pageSize:      10,
			expectedPages: 1,
		},
		{
			name:          "25 users, page 1, pageSize 10",
			users:         users,
			totalCount:    25,
			page:          1,
			pageSize:      10,
			expectedPages: 3,
		},
		{
			name:          "30 users, page 2, pageSize 10",
			users:         users,
			totalCount:    30,
			page:          2,
			pageSize:      10,
			expectedPages: 3,
		},
		{
			name:          "31 users, page 1, pageSize 10",
			users:         users,
			totalCount:    31,
			page:          1,
			pageSize:      10,
			expectedPages: 4,
		},
		{
			name:          "empty users list",
			users:         []UserResponse{},
			totalCount:    0,
			page:          1,
			pageSize:      10,
			expectedPages: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := NewPaginatedUsersResponse(tt.users, tt.totalCount, tt.page, tt.pageSize)

			if response == nil {
				t.Fatal("NewPaginatedUsersResponse() returned nil")
			}

			if len(response.Users) != len(tt.users) {
				t.Errorf("Expected %d users, got %d", len(tt.users), len(response.Users))
			}

			if response.TotalCount != tt.totalCount {
				t.Errorf("Expected TotalCount %d, got %d", tt.totalCount, response.TotalCount)
			}

			if response.Page != tt.page {
				t.Errorf("Expected Page %d, got %d", tt.page, response.Page)
			}

			if response.PageSize != tt.pageSize {
				t.Errorf("Expected PageSize %d, got %d", tt.pageSize, response.PageSize)
			}

			if response.TotalPages != tt.expectedPages {
				t.Errorf("Expected TotalPages %d, got %d", tt.expectedPages, response.TotalPages)
			}
		})
	}
}
