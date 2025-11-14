package model

import "testing"

func TestNewPaginatedOrdersResponse(t *testing.T) {
	orders := []Order{
		{ID: "o1", UserID: "user1"},
		{ID: "o2", UserID: "user2"},
		{ID: "o3", UserID: "user3"},
	}

	tests := []struct {
		name              string
		totalCount        int64
		page              int
		pageSize          int
		expectedTotalPages int
	}{
		{
			name:              "exact division",
			totalCount:        20,
			page:              1,
			pageSize:          10,
			expectedTotalPages: 2,
		},
		{
			name:              "with remainder",
			totalCount:        25,
			page:              1,
			pageSize:          10,
			expectedTotalPages: 3,
		},
		{
			name:              "single page",
			totalCount:        5,
			page:              1,
			pageSize:          10,
			expectedTotalPages: 1,
		},
		{
			name:              "zero items",
			totalCount:        0,
			page:              1,
			pageSize:          10,
			expectedTotalPages: 0,
		},
		{
			name:              "page size 1",
			totalCount:        100,
			page:              50,
			pageSize:          1,
			expectedTotalPages: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewPaginatedOrdersResponse(orders, tt.totalCount, tt.page, tt.pageSize)

			if result.Page != tt.page {
				t.Errorf("Page = %d, want %d", result.Page, tt.page)
			}
			if result.PageSize != tt.pageSize {
				t.Errorf("PageSize = %d, want %d", result.PageSize, tt.pageSize)
			}
			if result.TotalCount != tt.totalCount {
				t.Errorf("TotalCount = %d, want %d", result.TotalCount, tt.totalCount)
			}
			if result.TotalPages != tt.expectedTotalPages {
				t.Errorf("TotalPages = %d, want %d", result.TotalPages, tt.expectedTotalPages)
			}
			if len(result.Orders) != len(orders) {
				t.Errorf("Orders length = %d, want %d", len(result.Orders), len(orders))
			}
		})
	}
}

func TestNewPaginatedOrdersResponse_EmptyOrders(t *testing.T) {
	result := NewPaginatedOrdersResponse([]Order{}, 0, 1, 10)

	if result.TotalCount != 0 {
		t.Errorf("TotalCount = %d, want 0", result.TotalCount)
	}
	if result.TotalPages != 0 {
		t.Errorf("TotalPages = %d, want 0", result.TotalPages)
	}
	if len(result.Orders) != 0 {
		t.Errorf("Orders length = %d, want 0", len(result.Orders))
	}
}

func TestNewPaginatedOrdersResponse_NilOrders(t *testing.T) {
	result := NewPaginatedOrdersResponse(nil, 10, 1, 10)

	if result == nil {
		t.Error("Response should not be nil")
	}
	if result.TotalCount != 10 {
		t.Errorf("TotalCount = %d, want 10", result.TotalCount)
	}
	if result.Page != 1 {
		t.Errorf("Page = %d, want 1", result.Page)
	}
}
