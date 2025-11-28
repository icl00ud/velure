package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/icl00ud/publish-order-service/internal/middleware"
	"github.com/icl00ud/publish-order-service/internal/model"
	"github.com/icl00ud/publish-order-service/internal/service"
)

type fakeRepo struct {
	saveErr        error
	findErr        error
	foundOrder     model.Order
	paginated      *model.PaginatedOrdersResponse
	savedOrders    []model.Order
	findByUserErr  error
	getPageErr     error
	getPageUserErr error
}

func (f *fakeRepo) Save(ctx context.Context, o model.Order) error {
	f.savedOrders = append(f.savedOrders, o)
	return f.saveErr
}

func (f *fakeRepo) Find(ctx context.Context, id string) (model.Order, error) {
	if f.findErr != nil {
		return model.Order{}, f.findErr
	}
	return f.foundOrder, nil
}

func (f *fakeRepo) FindByUserID(ctx context.Context, userID, orderID string) (model.Order, error) {
	if f.findByUserErr != nil {
		return model.Order{}, f.findByUserErr
	}
	return f.foundOrder, nil
}

func (f *fakeRepo) GetOrdersByPage(ctx context.Context, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
	if f.getPageErr != nil {
		return nil, f.getPageErr
	}
	return f.paginated, nil
}

func (f *fakeRepo) GetOrdersByUserID(ctx context.Context, userID string, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
	if f.getPageUserErr != nil {
		return nil, f.getPageUserErr
	}
	return f.paginated, nil
}

func (f *fakeRepo) GetOrdersCount(ctx context.Context) (int64, error) {
	return 0, nil
}

func (f *fakeRepo) GetOrdersCountByUserID(ctx context.Context, userID string) (int64, error) {
	return 0, nil
}

type fakePricing struct {
	value float64
}

func (f fakePricing) Calculate(items []model.CartItem) float64 {
	return f.value
}

type fakePublisher struct {
	events []model.Event
	err    error
}

func (f *fakePublisher) Publish(evt model.Event) error {
	f.events = append(f.events, evt)
	return f.err
}

func newTestHandler(repo *fakeRepo, pub *fakePublisher, pricingValue float64) *OrderHandler {
	svc := service.NewOrderService(repo, fakePricing{value: pricingValue})
	return NewOrderHandler(svc, pub)
}

func withUser(ctx context.Context) context.Context {
	return context.WithValue(ctx, middleware.UserIDKey, "user-123")
}

func TestCreateOrder_Success(t *testing.T) {
	repo := &fakeRepo{}
	pub := &fakePublisher{}
	h := newTestHandler(repo, pub, 42.0)

	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"items":[{"product_id":"p1","quantity":2,"price":10}]}`))
	req = req.WithContext(withUser(req.Context()))
	w := httptest.NewRecorder()

	h.CreateOrder(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
	if len(repo.savedOrders) != 1 {
		t.Fatalf("expected Save to be called once, got %d", len(repo.savedOrders))
	}
	if repo.savedOrders[0].UserID != "user-123" {
		t.Fatalf("expected user id to be propagated, got %s", repo.savedOrders[0].UserID)
	}
	if repo.savedOrders[0].Total != 42.0 {
		t.Fatalf("expected total from pricing to be 42.0, got %f", repo.savedOrders[0].Total)
	}
	if len(pub.events) != 1 {
		t.Fatalf("expected one publish call, got %d", len(pub.events))
	}
	if pub.events[0].Type != model.OrderCreated {
		t.Fatalf("expected event type %s, got %s", model.OrderCreated, pub.events[0].Type)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response body: %v", err)
	}
	if resp["order_id"] == "" {
		t.Fatalf("expected order_id in response, got %v", resp)
	}
}

func TestCreateOrder_MissingUser(t *testing.T) {
	repo := &fakeRepo{}
	pub := &fakePublisher{}
	h := newTestHandler(repo, pub, 10)

	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`[]`))
	w := httptest.NewRecorder()

	h.CreateOrder(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestCreateOrder_InvalidPayload(t *testing.T) {
	repo := &fakeRepo{}
	pub := &fakePublisher{}
	h := newTestHandler(repo, pub, 10)

	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`invalid`))
	req = req.WithContext(withUser(req.Context()))
	w := httptest.NewRecorder()

	h.CreateOrder(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateOrder_ServiceError(t *testing.T) {
	repo := &fakeRepo{saveErr: errors.New("db down")}
	pub := &fakePublisher{}
	h := newTestHandler(repo, pub, 10)

	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"items":[{"product_id":"p1","quantity":1,"price":1}]}`))
	req = req.WithContext(withUser(req.Context()))
	w := httptest.NewRecorder()

	h.CreateOrder(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if len(pub.events) != 0 {
		t.Fatalf("expected no publish on failure, got %d events", len(pub.events))
	}
}

func TestCreateOrder_PublishErrorStillReturnsCreated(t *testing.T) {
	repo := &fakeRepo{}
	pub := &fakePublisher{err: errors.New("publish failed")}
	h := newTestHandler(repo, pub, 5)

	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"items":[{"product_id":"p1","quantity":1,"price":2}]}`))
	req = req.WithContext(withUser(req.Context()))
	w := httptest.NewRecorder()

	h.CreateOrder(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 even when publish fails, got %d", w.Code)
	}
	if len(pub.events) != 1 {
		t.Fatalf("expected publish to be attempted once, got %d", len(pub.events))
	}
}

func TestUpdateStatus_PublishesEvent(t *testing.T) {
	now := time.Now()
	repo := &fakeRepo{
		foundOrder: model.Order{ID: "order-1", UserID: "user-123", Status: model.StatusCreated, UpdatedAt: now},
	}
	pub := &fakePublisher{}
	h := newTestHandler(repo, pub, 0)

	req := httptest.NewRequest(http.MethodPost, "/update-order-status", strings.NewReader(`{"order_id":"order-1","status":"PROCESSING"}`))
	w := httptest.NewRecorder()

	h.UpdateStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if len(repo.savedOrders) != 1 {
		t.Fatalf("expected Save called once, got %d", len(repo.savedOrders))
	}
	if repo.savedOrders[0].Status != "PROCESSING" {
		t.Fatalf("expected status updated to PROCESSING, got %s", repo.savedOrders[0].Status)
	}
	if len(pub.events) != 1 {
		t.Fatalf("expected one publish call, got %d", len(pub.events))
	}
	if pub.events[0].Type != "PROCESSING" {
		t.Fatalf("expected event type PROCESSING, got %s", pub.events[0].Type)
	}
}

func TestUpdateStatus_InvalidPayload(t *testing.T) {
	repo := &fakeRepo{}
	pub := &fakePublisher{}
	h := newTestHandler(repo, pub, 0)

	req := httptest.NewRequest(http.MethodPost, "/update-order-status", strings.NewReader(`not-json`))
	w := httptest.NewRecorder()

	h.UpdateStatus(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid payload, got %d", w.Code)
	}
}

func TestUpdateStatus_FindError(t *testing.T) {
	repo := &fakeRepo{findErr: errors.New("not found")}
	pub := &fakePublisher{}
	h := newTestHandler(repo, pub, 0)

	req := httptest.NewRequest(http.MethodPost, "/update-order-status", strings.NewReader(`{"order_id":"order-1","status":"PROCESSING"}`))
	w := httptest.NewRecorder()

	h.UpdateStatus(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when find fails, got %d", w.Code)
	}
	if len(pub.events) != 0 {
		t.Fatalf("expected no publish on error, got %d", len(pub.events))
	}
}

func TestGetUserOrderByID_ValidatesInput(t *testing.T) {
	repo := &fakeRepo{}
	pub := &fakePublisher{}
	h := newTestHandler(repo, pub, 0)

	req := httptest.NewRequest(http.MethodGet, "/user/order", nil)
	req = req.WithContext(withUser(req.Context()))
	w := httptest.NewRecorder()

	h.GetUserOrderByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing id, got %d", w.Code)
	}
}

func TestGetUserOrderByID_NotFound(t *testing.T) {
	repo := &fakeRepo{findByUserErr: errors.New("missing")}
	pub := &fakePublisher{}
	h := newTestHandler(repo, pub, 0)

	req := httptest.NewRequest(http.MethodGet, "/user/order?id=order-123", nil)
	req = req.WithContext(withUser(req.Context()))
	w := httptest.NewRecorder()

	h.GetUserOrderByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when repo returns error, got %d", w.Code)
	}
}

func TestGetUserOrders_Unauthorized(t *testing.T) {
	repo := &fakeRepo{}
	pub := &fakePublisher{}
	h := newTestHandler(repo, pub, 0)

	req := httptest.NewRequest(http.MethodGet, "/user/orders", nil)
	w := httptest.NewRecorder()

	h.GetUserOrders(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when user is missing, got %d", w.Code)
	}
}

func TestGetUserOrders_RepoError(t *testing.T) {
	repo := &fakeRepo{getPageUserErr: errors.New("db issue")}
	pub := &fakePublisher{}
	h := newTestHandler(repo, pub, 0)

	req := httptest.NewRequest(http.MethodGet, "/user/orders", nil)
	req = req.WithContext(withUser(req.Context()))
	w := httptest.NewRecorder()

	h.GetUserOrders(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 on repo error, got %d", w.Code)
	}
}

func TestGetOrdersByPage_Success(t *testing.T) {
	repo := &fakeRepo{
		paginated: &model.PaginatedOrdersResponse{
			Orders:     []model.Order{{ID: "o1"}, {ID: "o2"}},
			TotalCount: 2,
			Page:       1,
			PageSize:   10,
			TotalPages: 1,
		},
	}
	pub := &fakePublisher{}
	h := newTestHandler(repo, pub, 0)

	req := httptest.NewRequest(http.MethodGet, "/orders?page=1&pageSize=10", nil)
	w := httptest.NewRecorder()

	h.GetOrdersByPage(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp model.PaginatedOrdersResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(resp.Orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(resp.Orders))
	}
}

func TestGetOrdersByPage_Error(t *testing.T) {
	repo := &fakeRepo{getPageErr: errors.New("db down")}
	pub := &fakePublisher{}
	h := newTestHandler(repo, pub, 0)

	req := httptest.NewRequest(http.MethodGet, "/orders?page=1&pageSize=10", nil)
	w := httptest.NewRecorder()

	h.GetOrdersByPage(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 on repo error, got %d", w.Code)
	}
}
