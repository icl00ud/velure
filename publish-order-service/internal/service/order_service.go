package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/icl00ud/publish-order-service/internal/model"
	"github.com/icl00ud/publish-order-service/internal/repository"
)

var ErrNoItems = errors.New("no items in the cart")

type OrderService struct {
	repo    repository.OrderRepository
	pricing PricingCalculator
}

func NewOrderService(r repository.OrderRepository, pc PricingCalculator) *OrderService {
	return &OrderService{repo: r, pricing: pc}
}

func (s *OrderService) Create(ctx context.Context, items []model.CartItem) (model.Order, error) {
	if len(items) == 0 {
		return model.Order{}, ErrNoItems
	}
	total := s.pricing.Calculate(items)
	now := time.Now()
	o := model.Order{
		ID:        uuid.NewString(),
		Items:     items,
		Total:     total,
		Status:    model.OrderCreated,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Save(ctx, o); err != nil {
		return model.Order{}, err
	}
	return o, nil
}

func (s *OrderService) UpdateStatus(ctx context.Context, id, status string) (model.Order, error) {
	o, err := s.repo.Find(ctx, id)
	if err != nil {
		return model.Order{}, err
	}
	o.Status = status
	o.UpdatedAt = time.Now()
	if err := s.repo.Save(ctx, o); err != nil {
		return model.Order{}, err
	}
	return o, nil
}

func (s *OrderService) GetOrdersByPage(ctx context.Context, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
	return s.repo.GetOrdersByPage(ctx, page, pageSize)
}
