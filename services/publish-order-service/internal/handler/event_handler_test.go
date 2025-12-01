package handler

import (
	"context"
	"testing"

	"github.com/icl00ud/publish-order-service/internal/model"
	"github.com/icl00ud/publish-order-service/internal/service"
	"github.com/icl00ud/velure-shared/logger"
)

type recordingRepo struct {
	findOrder model.Order
	findErr   error
	saved     []model.Order
}

func (r *recordingRepo) Save(ctx context.Context, order model.Order) error {
	r.saved = append(r.saved, order)
	return nil
}

func (r *recordingRepo) Find(ctx context.Context, id string) (model.Order, error) {
	if r.findErr != nil {
		return model.Order{}, r.findErr
	}
	order := r.findOrder
	order.ID = id
	return order, nil
}

func (r *recordingRepo) FindByUserID(ctx context.Context, userID, orderID string) (model.Order, error) {
	return model.Order{}, nil
}

func (r *recordingRepo) GetOrdersByPage(ctx context.Context, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
	return &model.PaginatedOrdersResponse{}, nil
}

func (r *recordingRepo) GetOrdersByUserID(ctx context.Context, userID string, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
	return &model.PaginatedOrdersResponse{}, nil
}

func (r *recordingRepo) GetOrdersCount(ctx context.Context) (int64, error) {
	return 0, nil
}

func (r *recordingRepo) GetOrdersCountByUserID(ctx context.Context, userID string) (int64, error) {
	return 0, nil
}

type fixedPricing struct{}

func (fixedPricing) Calculate(items []model.CartItem) float64 { return 0 }

func TestHandleEvent_UpdatesStatusAndNotifiesSSE(t *testing.T) {
	repo := &recordingRepo{findOrder: model.Order{UserID: "user-1", Status: model.StatusCreated}}
	svc := service.NewOrderService(repo, fixedPricing{})
	h := NewEventHandler(svc, logger.NewNop())

	sse := NewSSEHandler(svc)
	events := make(chan model.Order, 1)
	sse.registry.Register("order-1", events)
	h.SetSSEHandler(sse)

	err := h.HandleEvent(context.Background(), model.Event{
		Type:    model.OrderProcessing,
		Payload: []byte(`{"id":"order-1"}`),
	})
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}

	if len(repo.saved) == 0 || repo.saved[0].Status != model.StatusProcessing {
		t.Fatalf("expected order status to be updated to PROCESSING, got %+v", repo.saved)
	}

	select {
	case updated := <-events:
		if updated.Status != model.StatusProcessing {
			t.Fatalf("expected SSE to receive updated status, got %s", updated.Status)
		}
	default:
		t.Fatal("expected SSE notification to be sent")
	}
}

func TestHandleEvent_ReturnsErrorOnEmptyPayloadID(t *testing.T) {
	repo := &recordingRepo{findOrder: model.Order{UserID: "user-1"}}
	svc := service.NewOrderService(repo, fixedPricing{})
	h := NewEventHandler(svc, logger.NewNop())

	err := h.HandleEvent(context.Background(), model.Event{
		Type:    model.OrderCompleted,
		Payload: []byte(`{}`),
	})
	if err == nil {
		t.Fatal("expected error when payload misses order id")
	}
}

func TestHandleOrderProcessing_ValidPayload(t *testing.T) {
	repo := &recordingRepo{findOrder: model.Order{UserID: "user-1"}}
	svc := service.NewOrderService(repo, fixedPricing{})
	h := NewEventHandler(svc, logger.NewNop())

	if err := h.handleOrderProcessing(context.Background(), []byte(`{"id":"order-42"}`)); err != nil {
		t.Fatalf("expected no error updating status: %v", err)
	}
	if len(repo.saved) == 0 || repo.saved[0].Status != model.StatusProcessing || repo.saved[0].ID != "order-42" {
		t.Fatalf("expected status PROCESSING for order-42, got %+v", repo.saved)
	}
}

func TestHandleOrderProcessing_InvalidJSON(t *testing.T) {
	repo := &recordingRepo{findOrder: model.Order{}}
	svc := service.NewOrderService(repo, fixedPricing{})
	h := NewEventHandler(svc, logger.NewNop())

	if err := h.handleOrderProcessing(context.Background(), []byte(`{invalid`)); err == nil {
		t.Fatal("expected error for invalid JSON payload")
	}
}

func TestHandleOrderCompleted_UsesOrderIDField(t *testing.T) {
	repo := &recordingRepo{findOrder: model.Order{UserID: "user-1"}}
	svc := service.NewOrderService(repo, fixedPricing{})
	h := NewEventHandler(svc, logger.NewNop())

	if err := h.handleOrderCompleted(context.Background(), []byte(`{"order_id":"abc-123"}`)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.saved[0].ID != "abc-123" || repo.saved[0].Status != model.StatusCompleted {
		t.Fatalf("expected completed status for abc-123, got %+v", repo.saved[0])
	}
}
