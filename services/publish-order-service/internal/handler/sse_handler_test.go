package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/icl00ud/publish-order-service/internal/middleware"
	"github.com/icl00ud/publish-order-service/internal/model"
)

type fakeOrderService struct {
	order model.Order
	err   error
}

func (f *fakeOrderService) Create(ctx context.Context, userID string, items []model.CartItem) (model.Order, error) {
	return f.order, f.err
}
func (f *fakeOrderService) UpdateStatus(ctx context.Context, id, status string) (model.Order, error) {
	return f.order, f.err
}
func (f *fakeOrderService) GetOrdersByPage(ctx context.Context, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
	return &model.PaginatedOrdersResponse{}, f.err
}
func (f *fakeOrderService) GetOrdersByUserID(ctx context.Context, userID string, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
	return &model.PaginatedOrdersResponse{}, f.err
}
func (f *fakeOrderService) GetOrderByID(ctx context.Context, userID, orderID string) (model.Order, error) {
	return f.order, f.err
}
func (f *fakeOrderService) GetUserOrders(ctx context.Context, userID string) ([]model.Order, error) {
	return nil, f.err
}
func (f *fakeOrderService) GetUserOrderByID(ctx context.Context, userID, orderID string) (model.Order, error) {
	return f.order, f.err
}

func TestStreamOrderStatus_Unauthorized(t *testing.T) {
	handler := NewSSEHandler(&fakeOrderService{})

	req := httptest.NewRequest(http.MethodGet, "/user/order/status?id=o1", nil)
	rr := httptest.NewRecorder()

	handler.StreamOrderStatus(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestStreamOrderStatus_MissingOrderID(t *testing.T) {
	handler := NewSSEHandler(&fakeOrderService{})

	req := httptest.NewRequest(http.MethodGet, "/user/order/status", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user1"))
	rr := httptest.NewRecorder()

	handler.StreamOrderStatus(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestStreamOrderStatus_OrderNotFound(t *testing.T) {
	svc := &fakeOrderService{err: errors.New("not found")}
	handler := &SSEHandler{svc: svc, registry: NewSSERegistry()}

	req := httptest.NewRequest(http.MethodGet, "/user/order/status?id=o1", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user1"))
	rr := httptest.NewRecorder()

	handler.StreamOrderStatus(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestStreamOrderStatus_SendsInitialAndKeepalive(t *testing.T) {
	svc := &fakeOrderService{order: model.Order{ID: "o1", Status: model.StatusProcessing}}
	handler := &SSEHandler{svc: svc, registry: NewSSERegistry()}

	req := httptest.NewRequest(http.MethodGet, "/user/order/status?id=o1", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user1"))
	rr := httptest.NewRecorder()

	// use close notifier context to cancel quickly
	ctx, cancel := context.WithTimeout(req.Context(), 50*time.Millisecond)
	defer cancel()
	req = req.WithContext(ctx)

	go handler.StreamOrderStatus(rr, req)
	time.Sleep(20 * time.Millisecond)
	cancel()

	body := rr.Body.String()
	if !strings.Contains(body, `"status":"PROCESSING"`) {
		t.Fatalf("expected initial event in body, got %s", body)
	}
}

func TestSSERegistry_RegisterAndBroadcast(t *testing.T) {
	reg := NewSSERegistry()
	ch := make(chan model.Order, 1)
	reg.Register("o1", ch)

	order := model.Order{ID: "o1", Status: model.StatusCompleted}
	reg.Broadcast("o1", order)

	select {
	case got := <-ch:
		if got.Status != model.StatusCompleted {
			t.Fatalf("expected status %s, got %s", model.StatusCompleted, got.Status)
		}
	default:
		t.Fatal("expected broadcast to deliver message")
	}

	reg.Unregister("o1", ch)
	if _, ok := reg.subscribers["o1"]; ok {
		t.Fatal("expected subscriber removed after unregister")
	}
}
