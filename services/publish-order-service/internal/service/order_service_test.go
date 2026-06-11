package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"github.com/icl00ud/velure/services/publish-order-service/internal/model"
	"github.com/icl00ud/velure/services/publish-order-service/internal/outbox"
	"github.com/icl00ud/velure/services/publish-order-service/internal/repository"
)

// Mock repository for testing
type mockOrderRepository struct {
	saveFunc                   func(ctx context.Context, order model.Order) error
	findFunc                   func(ctx context.Context, id string) (model.Order, error)
	findByUserIDFunc           func(ctx context.Context, userID, orderID string) (model.Order, error)
	getOrdersByPageFunc        func(ctx context.Context, page, pageSize int) (*model.PaginatedOrdersResponse, error)
	getOrdersByUserIDFunc      func(ctx context.Context, userID string, page, pageSize int) (*model.PaginatedOrdersResponse, error)
	getOrdersCountFunc         func(ctx context.Context) (int64, error)
	getOrdersCountByUserIDFunc func(ctx context.Context, userID string) (int64, error)
}

func (m *mockOrderRepository) Save(ctx context.Context, order model.Order) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, order)
	}
	return nil
}

func (m *mockOrderRepository) SaveTx(ctx context.Context, _ *sql.Tx, order model.Order) error {
	return m.Save(ctx, order)
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

// mockOutboxRepository is a no-op outbox for tests that use mockOrderRepository.
type mockOutboxRepository struct {
	saveTxErr error
}

func (m *mockOutboxRepository) SaveTx(_ context.Context, _ *sql.Tx, _ model.OutboxEvent) error {
	return m.saveTxErr
}

func (m *mockOutboxRepository) FetchUnpublished(_ context.Context, _ int) (*sql.Tx, []model.OutboxEvent, error) {
	return nil, nil, nil
}

func (m *mockOutboxRepository) MarkPublished(_ context.Context, _ *sql.Tx, _ []string) error {
	return nil
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

// newMockDB returns a sqlmock db pre-configured with Begin+Commit expectations,
// used by tests that exercise Create/UpdateStatus via mockOrderRepository (which
// intercepts SaveTx before any SQL is executed on the tx).
func newMockDBWithTx(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	return db, mock
}

func TestNewOrderService(t *testing.T) {
	repo := &mockOrderRepository{}
	pc := &mockPricingCalculator{}
	svc := NewOrderService(repo, nil, nil, pc)

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
			name:   "invalid item - missing product_id",
			userID: "user123",
			items: []model.CartItem{
				{ProductID: "", Name: "Product 1", Quantity: 1, Price: 15.0},
			},
			pricing:     0.0,
			saveErr:     nil,
			expectedErr: fmt.Errorf("%w: missing product_id", ErrInvalidItem),
		},
		{
			name:   "invalid item - non-positive quantity",
			userID: "user123",
			items: []model.CartItem{
				{ProductID: "p1", Name: "Product 1", Quantity: 0, Price: 15.0},
			},
			pricing:     0.0,
			saveErr:     nil,
			expectedErr: fmt.Errorf("%w: quantity must be positive", ErrInvalidItem),
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

			// Validation errors are returned before BeginTx; success/save-error paths need a tx.
			needsTx := tt.expectedErr == nil || tt.saveErr != nil
			var db *sql.DB
			var mock sqlmock.Sqlmock
			if needsTx {
				db, mock = newMockDBWithTx(t)
				defer db.Close()
				mock.ExpectBegin()
				if tt.saveErr != nil {
					mock.ExpectRollback()
				} else {
					mock.ExpectCommit()
				}
			}

			svc := NewOrderService(repo, &mockOutboxRepository{}, db, pc)
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

			if mock != nil {
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("sqlmock expectations: %v", err)
				}
			}
		})
	}
}

func TestOrderService_UpdateStatus(t *testing.T) {
	tests := []struct {
		name        string
		orderID     string
		newStatus   string
		findErr     error
		saveErr     error
		expectedErr error
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

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer db.Close()

			mock.ExpectBegin()
			if tt.expectedErr != nil {
				mock.ExpectRollback()
			} else {
				mock.ExpectCommit()
			}

			svc := NewOrderService(repo, &mockOutboxRepository{}, db, pc)
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

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock expectations: %v", err)
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
	svc := NewOrderService(repo, nil, nil, pc)

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
	svc := NewOrderService(repo, nil, nil, pc)

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
	svc := NewOrderService(repo, nil, nil, pc)

	order, err := svc.GetOrderByID(context.Background(), "user123", "order123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if order.ID != expectedOrder.ID {
		t.Errorf("expected order ID %s, got %s", expectedOrder.ID, order.ID)
	}
}

// ---------------------------------------------------------------------------
// Atomic outbox tests — use real repository impls over sqlmock db.
// ---------------------------------------------------------------------------

func TestOrderService_Create_PersistsOrderAndOutboxAtomically(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO TBLOrders`).
		WithArgs(sqlmock.AnyArg(), "user-1", sqlmock.AnyArg(), float64(0), "CREATED", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO outbox_events`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), model.OrderCreated, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := repository.NewOrderRepositoryFromDB(db)
	outboxRepo := outbox.NewPostgresRepository(db)
	svc := NewOrderService(repo, outboxRepo, db, NewPricingCalculator())

	_, err = svc.Create(context.Background(), "user-1", []model.CartItem{{ProductID: "p1", Quantity: 1}})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestOrderService_Create_RollsBackOnOutboxFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO TBLOrders`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO outbox_events`).
		WillReturnError(errors.New("outbox down"))
	mock.ExpectRollback()

	repo := repository.NewOrderRepositoryFromDB(db)
	outboxRepo := outbox.NewPostgresRepository(db)
	svc := NewOrderService(repo, outboxRepo, db, NewPricingCalculator())

	_, err = svc.Create(context.Background(), "user-1", []model.CartItem{{ProductID: "p1", Quantity: 1}})
	if err == nil {
		t.Fatal("expected error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestOrderService_Create_CapturesTraceContextInOutbox(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO TBLOrders`).WillReturnResult(sqlmock.NewResult(0, 1))
	// traceparent format: 00-<32 hex>-<16 hex>-<2 hex flags>
	mock.ExpectExec(`INSERT INTO outbox_events`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), model.OrderCreated, sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := repository.NewOrderRepositoryFromDB(db)
	rec := &recordingOutbox{inner: outbox.NewPostgresRepository(db)}
	svc := NewOrderService(repo, rec, db, NewPricingCalculator())

	// Build a context with an active recording span.
	otel.SetTextMapPropagator(propagation.TraceContext{})
	tp := sdktrace.NewTracerProvider()
	defer tp.Shutdown(context.Background())
	ctx, span := tp.Tracer("test").Start(context.Background(), "create-order")
	defer span.End()

	if _, err := svc.Create(ctx, "user-1", []model.CartItem{{ProductID: "p1", Quantity: 1}}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	want := "00-" + span.SpanContext().TraceID().String()
	if rec.lastEvent.TraceContext == "" || !strings.HasPrefix(rec.lastEvent.TraceContext, want) {
		t.Fatalf("expected outbox TraceContext to carry trace id %s, got %q", want, rec.lastEvent.TraceContext)
	}
}

type recordingOutbox struct {
	inner     outbox.Repository
	lastEvent model.OutboxEvent
}

func (r *recordingOutbox) SaveTx(ctx context.Context, tx *sql.Tx, evt model.OutboxEvent) error {
	r.lastEvent = evt
	return r.inner.SaveTx(ctx, tx, evt)
}

func (r *recordingOutbox) FetchUnpublished(ctx context.Context, limit int) (*sql.Tx, []model.OutboxEvent, error) {
	return r.inner.FetchUnpublished(ctx, limit)
}

func (r *recordingOutbox) MarkPublished(ctx context.Context, tx *sql.Tx, ids []string) error {
	return r.inner.MarkPublished(ctx, tx, ids)
}
