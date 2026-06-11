package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/icl00ud/velure/services/publish-order-service/internal/middleware"
	"github.com/icl00ud/velure/services/publish-order-service/internal/model"
	"github.com/icl00ud/velure/services/publish-order-service/internal/service"
)

// fakeOutboxRepository is a no-op outbox used in handler tests.
type fakeOutboxRepository struct{}

func (f *fakeOutboxRepository) SaveTx(_ context.Context, _ *sql.Tx, _ model.OutboxEvent) error {
	return nil
}

func (f *fakeOutboxRepository) FetchUnpublished(_ context.Context, _ int) (*sql.Tx, []model.OutboxEvent, error) {
	return nil, nil, nil
}

func (f *fakeOutboxRepository) MarkPublished(_ context.Context, _ *sql.Tx, _ []string) error {
	return nil
}

// newPermissiveDB returns a sqlmock *sql.DB that accepts Begin+Commit and
// Begin+Rollback without strict expectations. Used by handler tests that
// exercise Create/UpdateStatus but aren't testing the tx itself.
func newPermissiveDB(t *testing.T) *sql.DB {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	mock.MatchExpectationsInOrder(false)
	// Pre-register enough Begin/Commit/Rollback expectations for handler tests
	// that may call Create or UpdateStatus multiple times.
	for i := 0; i < 5; i++ {
		mock.ExpectBegin()
		mock.ExpectCommit()
		mock.ExpectRollback()
	}
	t.Cleanup(func() { db.Close() })
	return db
}

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

func (f *fakeRepo) SaveTx(ctx context.Context, _ *sql.Tx, o model.Order) error {
	return f.Save(ctx, o)
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

func newTestHandler(t *testing.T, repo *fakeRepo, pricingValue float64) *OrderHandler {
	db := newPermissiveDB(t)
	svc := service.NewOrderService(repo, &fakeOutboxRepository{}, db, fakePricing{value: pricingValue})
	return NewOrderHandler(svc)
}

func withUser(ctx context.Context) context.Context {
	return context.WithValue(ctx, middleware.UserIDKey, "user-123")
}

func TestCreateOrder_Success(t *testing.T) {
	repo := &fakeRepo{}
	h := newTestHandler(t, repo, 42.0)

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
	h := newTestHandler(t, repo, 10)

	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`[]`))
	w := httptest.NewRecorder()

	h.CreateOrder(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestCreateOrder_InvalidPayload(t *testing.T) {
	repo := &fakeRepo{}
	h := newTestHandler(t, repo, 10)

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
	h := newTestHandler(t, repo, 10)

	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"items":[{"product_id":"p1","quantity":1,"price":1}]}`))
	req = req.WithContext(withUser(req.Context()))
	w := httptest.NewRecorder()

	h.CreateOrder(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "db down") {
		t.Fatalf("internal error leaked to client: %s", w.Body.String())
	}
}

func TestCreateOrder_OutboxWrittenReturnsCreated(t *testing.T) {
	repo := &fakeRepo{}
	h := newTestHandler(t, repo, 5)

	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"items":[{"product_id":"p1","quantity":1,"price":2}]}`))
	req = req.WithContext(withUser(req.Context()))
	w := httptest.NewRecorder()

	h.CreateOrder(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 after outbox write, got %d", w.Code)
	}
}

func TestCreateOrder_ReturnsCreatedAfterServiceWrite(t *testing.T) {
	repo := &fakeRepo{}
	h := newTestHandler(t, repo, 10)

	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"items":[{"product_id":"p1","quantity":1,"price":2}]}`))
	req = req.WithContext(withUser(req.Context()))
	w := httptest.NewRecorder()

	h.CreateOrder(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestCreateOrder_NoItemsReturnsBadRequest(t *testing.T) {
	repo := &fakeRepo{}
	h := newTestHandler(t, repo, 0)

	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"items":[]}`))
	req = req.WithContext(withUser(req.Context()))
	w := httptest.NewRecorder()

	h.CreateOrder(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for no items, got %d", w.Code)
	}
}

func TestCreateOrder_InvalidItem(t *testing.T) {
	repo := &fakeRepo{}
	h := newTestHandler(t, repo, 0)

	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"items":[{"product_id":"","quantity":0,"price":1}]}`))
	req = req.WithContext(withUser(req.Context()))
	w := httptest.NewRecorder()

	h.CreateOrder(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid item, got %d", w.Code)
	}
}

func TestGetUserOrderByID_ValidatesInput(t *testing.T) {
	repo := &fakeRepo{}
	h := newTestHandler(t, repo, 0)

	req := httptest.NewRequest(http.MethodGet, "/user/order", nil)
	req = req.WithContext(withUser(req.Context()))
	w := httptest.NewRecorder()

	h.GetUserOrderByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing id, got %d", w.Code)
	}
}

func TestGetUserOrderByID_Unauthorized(t *testing.T) {
	repo := &fakeRepo{}
	h := newTestHandler(t, repo, 0)

	req := httptest.NewRequest(http.MethodGet, "/user/order?id=order-123", nil)
	w := httptest.NewRecorder()

	h.GetUserOrderByID(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when user missing, got %d", w.Code)
	}
}

func TestGetUserOrderByID_NotFound(t *testing.T) {
	repo := &fakeRepo{findByUserErr: errors.New("missing")}
	h := newTestHandler(t, repo, 0)

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
	h := newTestHandler(t, repo, 0)

	req := httptest.NewRequest(http.MethodGet, "/user/orders", nil)
	w := httptest.NewRecorder()

	h.GetUserOrders(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when user is missing, got %d", w.Code)
	}
}

func TestGetUserOrders_RepoError(t *testing.T) {
	repo := &fakeRepo{getPageUserErr: errors.New("db issue")}
	h := newTestHandler(t, repo, 0)

	req := httptest.NewRequest(http.MethodGet, "/user/orders", nil)
	req = req.WithContext(withUser(req.Context()))
	w := httptest.NewRecorder()

	h.GetUserOrders(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 on repo error, got %d", w.Code)
	}
}

