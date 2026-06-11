package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/icl00ud/velure/services/publish-order-service/internal/model"
	"github.com/icl00ud/velure/services/publish-order-service/internal/outbox"
	"github.com/icl00ud/velure/services/publish-order-service/internal/repository"
	"github.com/icl00ud/velure/services/publish-order-service/internal/telemetry"
)

var ErrNoItems = errors.New("no items in the cart")
var ErrInvalidItem = errors.New("invalid item in the cart")

type OrderService struct {
	repo    repository.OrderRepository
	outbox  outbox.Repository
	db      *sql.DB
	pricing PricingCalculator
}

func NewOrderService(r repository.OrderRepository, ob outbox.Repository, db *sql.DB, pc PricingCalculator) *OrderService {
	return &OrderService{repo: r, outbox: ob, db: db, pricing: pc}
}

func (s *OrderService) Create(ctx context.Context, userID string, items []model.CartItem) (model.Order, error) {
	if len(items) == 0 {
		return model.Order{}, ErrNoItems
	}
	for _, item := range items {
		if item.ProductID == "" {
			return model.Order{}, fmt.Errorf("%w: missing product_id", ErrInvalidItem)
		}
		if item.Quantity <= 0 {
			return model.Order{}, fmt.Errorf("%w: quantity must be positive", ErrInvalidItem)
		}
	}

	total := s.pricing.Calculate(items)
	now := time.Now()
	o := model.Order{
		ID:        uuid.NewString(),
		UserID:    userID,
		Items:     items,
		Total:     total,
		Status:    model.StatusCreated,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.withTx(ctx, func(tx *sql.Tx) error {
		if err := s.repo.SaveTx(ctx, tx, o); err != nil {
			return err
		}
		payload, err := json.Marshal(o)
		if err != nil {
			return err
		}
		return s.outbox.SaveTx(ctx, tx, model.OutboxEvent{
			ID:           uuid.NewString(),
			AggregateID:  o.ID,
			EventType:    model.OrderCreated,
			Payload:      payload,
			CreatedAt:    now,
			TraceContext: telemetry.Traceparent(ctx),
		})
	}); err != nil {
		return model.Order{}, err
	}
	return o, nil
}

func (s *OrderService) UpdateStatus(ctx context.Context, id, status string) (model.Order, error) {
	var updated model.Order
	if err := s.withTx(ctx, func(tx *sql.Tx) error {
		o, err := s.repo.Find(ctx, id)
		if err != nil {
			return err
		}
		o.Status = status
		o.UpdatedAt = time.Now()
		if err := s.repo.SaveTx(ctx, tx, o); err != nil {
			return err
		}
		payload, err := json.Marshal(o)
		if err != nil {
			return err
		}
		if err := s.outbox.SaveTx(ctx, tx, model.OutboxEvent{
			ID:           uuid.NewString(),
			AggregateID:  o.ID,
			EventType:    status,
			Payload:      payload,
			CreatedAt:    o.UpdatedAt,
			TraceContext: telemetry.Traceparent(ctx),
		}); err != nil {
			return err
		}
		updated = o
		return nil
	}); err != nil {
		return model.Order{}, err
	}
	return updated, nil
}

func (s *OrderService) withTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *OrderService) GetOrdersByPage(ctx context.Context, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
	return s.repo.GetOrdersByPage(ctx, page, pageSize)
}

func (s *OrderService) GetOrdersByUserID(ctx context.Context, userID string, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
	return s.repo.GetOrdersByUserID(ctx, userID, page, pageSize)
}

func (s *OrderService) GetOrderByID(ctx context.Context, userID, orderID string) (model.Order, error) {
	return s.repo.FindByUserID(ctx, userID, orderID)
}
