package handler

import (
	"context"

	"github.com/icl00ud/publish-order-service/internal/model"
)

// OrderService defines the operations used by handlers.
type OrderService interface {
	Create(ctx context.Context, userID string, items []model.CartItem) (model.Order, error)
	UpdateStatus(ctx context.Context, id, status string) (model.Order, error)
	GetOrdersByPage(ctx context.Context, page, pageSize int) (*model.PaginatedOrdersResponse, error)
	GetOrdersByUserID(ctx context.Context, userID string, page, pageSize int) (*model.PaginatedOrdersResponse, error)
	GetOrderByID(ctx context.Context, userID, orderID string) (model.Order, error)
}
