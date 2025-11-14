package service

import (
	"context"
	"errors"
	"testing"

	"github.com/icl00ud/publish-order-service/internal/model"
)

// Mock repository for testing
type mockOrderRepository struct {
	saveFunc         func(ctx context.Context, order model.Order) error
	findFunc         func(ctx context.Context, id string) (model.Order, error)
	findByUserIDFunc func(ctx context.Context, userID, orderID string) (model.Order, error)
	getOrdersByPageFunc func(ctx context.Context, page, pageSize int) (*model.PaginatedOrdersResponse, error)
	getOrdersByUserIDFunc func(ctx context.Context, userID string, page, pageSize int) (*model.PaginatedOrdersResponse, error)
	getOrdersCountFunc func(ctx context.Context) (int64, error)
	getOrdersCountByUserIDFunc func(ctx context.Context, userID string) (int64, error)
}

func (m *mockOrderRepository) Save(ctx context.Context, order model.Order) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, order)
	}
	return nil
}

func (m *mockOrderRepository) Find(ctx context.Context, id string) (model.Order, error) {
	if m.findFunc != nil {
		return m.findFunc(ctx, id)
	}
	return model.Order{}, nil
}

func (m *mockOrderRepository) FindByUserID(ctx context.Context, userID, orderID string) (model.Order, error) {
	if m.findByUserIDFunc != nil {
		return m.findByUserIDFunc(ctx, userID, orderID)
	}
	return model.Order{}, nil
}

func (m *mockOrderRepository) GetOrdersByPage(ctx context.Context, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
	if m.getOrdersByPageFunc != nil {
		return m.getOrdersByPageFunc(ctx, page, pageSize)
	}
	return &model.PaginatedOrdersResponse{}, nil
}

func (m *mockOrderRepository) GetOrdersByUserID(ctx context.Context, userID string, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
	if m.getOrdersByUserIDFunc != nil {
		return m.getOrdersByUserIDFunc(ctx, userID, page, pageSize)
	}
	return &model.PaginatedOrdersResponse{}, nil
}

func (m *mockOrderRepository) GetOrdersCount(ctx context.Context) (int64, error) {
	if m.getOrdersCountFunc != nil {
		return m.getOrdersCountFunc(ctx)
	}
	return 0, nil
}

func (m *mockOrderRepository) GetOrdersCountByUserID(ctx context.Context, userID string) (int64, error) {
	if m.getOrdersCountByUserIDFunc != nil {
		return m.getOrdersCountByUserIDFunc(ctx, userID)
	}
	return 0, nil
}

// Mock pricing calculator
type mockPricingCalculator struct {
	calculateFunc func(items []model.CartItem) float64
}

func (m *mockPricingCalculator) Calculate(items []model.CartItem) float64 {
	if m.calculateFunc != nil {
		return m.calculateFunc(items)
	}
	return 0.0
}

func TestNewOrderService(t *testing.T) {
	repo := &mockOrderRepository{}
	pc := &mockPricingCalculator{}
	svc := NewOrderService(repo, pc)

	if svc == nil {
		t.Fatal("expected non-nil order service")
	}
}

func TestOrderService_Create(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		items         []model.CartItem
		pricing       float64
		saveErr       error
		expectedErr   error
		expectedTotal float64
	}{
		{
			name:   "successful order creation",
			userID: "user123",
			items: []model.CartItem{
				{ProductID: "p1", Name: "Product 1", Quantity: 2, Price: 10.0},
			},
			pricing:       20.0,
			saveErr:       nil,
			expectedErr:   nil,
			expectedTotal: 20.0,
		},
		{
			name:        "empty items",
			userID:      "user123",
			items:       []model.CartItem{},
			pricing:     0.0,
			saveErr:     nil,
			expectedErr: ErrNoItems,
		},
		{
			name:   "repository save error",
			userID: "user123",
			items: []model.CartItem{
				{ProductID: "p1", Name: "Product 1", Quantity: 1, Price: 15.0},
			},
			pricing:     15.0,
			saveErr:     errors.New("database error"),
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOrderRepository{
				saveFunc: func(ctx context.Context, order model.Order) error {
					return tt.saveErr
				},
			}
			pc := &mockPricingCalculator{
				calculateFunc: func(items []model.CartItem) float64 {
					return tt.pricing
				},
			}
			svc := NewOrderService(repo, pc)

			order, err := svc.Create(context.Background(), tt.userID, tt.items)

			if tt.expectedErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedErr)
				} else if err.Error() != tt.expectedErr.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if order.UserID != tt.userID {
					t.Errorf("expected userID %s, got %s", tt.userID, order.UserID)
				}
				if order.Total != tt.expectedTotal {
					t.Errorf("expected total %f, got %f", tt.expectedTotal, order.Total)
				}
				if order.Status != model.StatusCreated {
					t.Errorf("expected status %s, got %s", model.StatusCreated, order.Status)
				}
				if order.ID == "" {
					t.Error("expected non-empty order ID")
				}
			}
		})
	}
}

func TestOrderService_UpdateStatus(t *testing.T) {
	tests := []struct {
		name         string
		orderID      string
		newStatus    string
		findErr      error
		saveErr      error
		expectedErr  error
	}{
		{
			name:        "successful status update",
			orderID:     "order123",
			newStatus:   model.StatusProcessing,
			findErr:     nil,
			saveErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "order not found",
			orderID:     "nonexistent",
			newStatus:   model.StatusProcessing,
			findErr:     errors.New("order not found"),
			saveErr:     nil,
			expectedErr: errors.New("order not found"),
		},
		{
			name:        "save error",
			orderID:     "order123",
			newStatus:   model.StatusCompleted,
			findErr:     nil,
			saveErr:     errors.New("database error"),
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOrderRepository{
				findFunc: func(ctx context.Context, id string) (model.Order, error) {
					if tt.findErr != nil {
						return model.Order{}, tt.findErr
					}
					return model.Order{
						ID:     tt.orderID,
						UserID: "user123",
						Status: model.StatusCreated,
					}, nil
				},
				saveFunc: func(ctx context.Context, order model.Order) error {
					return tt.saveErr
				},
			}
			pc := &mockPricingCalculator{}
			svc := NewOrderService(repo, pc)

			order, err := svc.UpdateStatus(context.Background(), tt.orderID, tt.newStatus)

			if tt.expectedErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedErr)
				} else if err.Error() != tt.expectedErr.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if order.Status != tt.newStatus {
					t.Errorf("expected status %s, got %s", tt.newStatus, order.Status)
				}
			}
		})
	}
}

func TestOrderService_GetOrdersByPage(t *testing.T) {
	expectedResponse := &model.PaginatedOrdersResponse{
		Orders: []model.Order{
			{ID: "o1", UserID: "user1"},
			{ID: "o2", UserID: "user2"},
		},
		Page:       1,
		PageSize:   10,
		TotalCount: 2,
	}

	repo := &mockOrderRepository{
		getOrdersByPageFunc: func(ctx context.Context, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
			return expectedResponse, nil
		},
	}
	pc := &mockPricingCalculator{}
	svc := NewOrderService(repo, pc)

	result, err := svc.GetOrdersByPage(context.Background(), 1, 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.TotalCount != expectedResponse.TotalCount {
		t.Errorf("expected total %d, got %d", expectedResponse.TotalCount, result.TotalCount)
	}
}

func TestOrderService_GetOrdersByUserID(t *testing.T) {
	expectedResponse := &model.PaginatedOrdersResponse{
		Orders: []model.Order{
			{ID: "o1", UserID: "user123"},
		},
		Page:       1,
		PageSize:   10,
		TotalCount: 1,
	}

	repo := &mockOrderRepository{
		getOrdersByUserIDFunc: func(ctx context.Context, userID string, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
			return expectedResponse, nil
		},
	}
	pc := &mockPricingCalculator{}
	svc := NewOrderService(repo, pc)

	result, err := svc.GetOrdersByUserID(context.Background(), "user123", 1, 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.TotalCount != expectedResponse.TotalCount {
		t.Errorf("expected total %d, got %d", expectedResponse.TotalCount, result.TotalCount)
	}
}

func TestOrderService_GetOrderByID(t *testing.T) {
	expectedOrder := model.Order{
		ID:     "order123",
		UserID: "user123",
		Status: model.StatusCreated,
	}

	repo := &mockOrderRepository{
		findByUserIDFunc: func(ctx context.Context, userID, orderID string) (model.Order, error) {
			if orderID == "order123" && userID == "user123" {
				return expectedOrder, nil
			}
			return model.Order{}, errors.New("not found")
		},
	}
	pc := &mockPricingCalculator{}
	svc := NewOrderService(repo, pc)

	order, err := svc.GetOrderByID(context.Background(), "user123", "order123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if order.ID != expectedOrder.ID {
		t.Errorf("expected order ID %s, got %s", expectedOrder.ID, order.ID)
	}
}
