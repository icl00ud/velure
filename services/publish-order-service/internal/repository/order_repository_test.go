// repository/order_repository_unit_test.go
package repository

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/icl00ud/publish-order-service/internal/model"
)

func TestPostgresOrderRepository_Save(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("erro sqlmock: %v", err)
	}
	defer db.Close()

	repo := &PostgresOrderRepository{db: db}

	now := time.Now().Truncate(time.Second)
	items := []model.CartItem{{ProductID: "p1", Name: "n1", Quantity: 2, Price: 10.0}}
	order := model.Order{
		ID:        "o1",
		UserID:    "user123",
		Items:     items,
		Total:     20,
		Status:    model.OrderCreated,
		CreatedAt: now,
		UpdatedAt: now,
	}

	itemsJSON, _ := json.Marshal(items)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO TBLOrders")).
		WithArgs(order.ID, order.UserID, itemsJSON, order.Total, order.Status, order.CreatedAt, order.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.Save(context.Background(), order); err != nil {
		t.Errorf("Save retornou erro: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations não atendidas: %v", err)
	}
}

func TestPostgresOrderRepository_Find(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("erro sqlmock: %v", err)
	}
	defer db.Close()

	repo := &PostgresOrderRepository{db: db}

	now := time.Now().Truncate(time.Second)
	items := []model.CartItem{{ProductID: "p1", Name: "n1", Quantity: 2, Price: 10.0}}
	itemsJSON, _ := json.Marshal(items)
	order := model.Order{
		ID:        "o2",
		UserID:    "user123",
		Items:     items,
		Total:     20,
		Status:    model.OrderCreated,
		CreatedAt: now,
		UpdatedAt: now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "items", "total", "status", "created_at", "updated_at",
	}).AddRow(
		order.ID, order.UserID, itemsJSON, order.Total, order.Status, order.CreatedAt, order.UpdatedAt,
	)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, user_id, items, total, status, created_at, updated_at")).
		WithArgs(order.ID).
		WillReturnRows(rows)

	got, err := repo.Find(context.Background(), order.ID)
	if err != nil {
		t.Errorf("Find retornou erro: %v", err)
	}
	if got.ID != order.ID || got.Total != order.Total || got.Status != order.Status {
		t.Errorf("Find retornou %+v; want %+v", got, order)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations não atendidas: %v", err)
	}
}

func TestPostgresOrderRepository_FindByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("erro sqlmock: %v", err)
	}
	defer db.Close()

	repo := &PostgresOrderRepository{db: db}

	now := time.Now().Truncate(time.Second)
	items := []model.CartItem{{ProductID: "p1", Name: "n1", Quantity: 2, Price: 10.0}}
	itemsJSON, _ := json.Marshal(items)
	order := model.Order{
		ID:        "o3",
		UserID:    "user456",
		Items:     items,
		Total:     20,
		Status:    model.OrderCreated,
		CreatedAt: now,
		UpdatedAt: now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "items", "total", "status", "created_at", "updated_at",
	}).AddRow(
		order.ID, order.UserID, itemsJSON, order.Total, order.Status, order.CreatedAt, order.UpdatedAt,
	)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, user_id, items, total, status, created_at, updated_at")).
		WithArgs(order.ID, order.UserID).
		WillReturnRows(rows)

	got, err := repo.FindByUserID(context.Background(), order.UserID, order.ID)
	if err != nil {
		t.Errorf("FindByUserID retornou erro: %v", err)
	}
	if got.ID != order.ID || got.UserID != order.UserID {
		t.Errorf("FindByUserID retornou %+v; want %+v", got, order)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations não atendidas: %v", err)
	}
}

func TestPostgresOrderRepository_GetOrdersCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("erro sqlmock: %v", err)
	}
	defer db.Close()

	repo := &PostgresOrderRepository{db: db}

	rows := sqlmock.NewRows([]string{"count"}).AddRow(42)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM TBLOrders")).
		WillReturnRows(rows)

	count, err := repo.GetOrdersCount(context.Background())
	if err != nil {
		t.Errorf("GetOrdersCount retornou erro: %v", err)
	}
	if count != 42 {
		t.Errorf("GetOrdersCount retornou %d; want 42", count)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations não atendidas: %v", err)
	}
}

func TestPostgresOrderRepository_GetOrdersCountByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("erro sqlmock: %v", err)
	}
	defer db.Close()

	repo := &PostgresOrderRepository{db: db}

	rows := sqlmock.NewRows([]string{"count"}).AddRow(5)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM TBLOrders WHERE user_id = $1")).
		WithArgs("user789").
		WillReturnRows(rows)

	count, err := repo.GetOrdersCountByUserID(context.Background(), "user789")
	if err != nil {
		t.Errorf("GetOrdersCountByUserID retornou erro: %v", err)
	}
	if count != 5 {
		t.Errorf("GetOrdersCountByUserID retornou %d; want 5", count)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations não atendidas: %v", err)
	}
}

func TestPostgresOrderRepository_GetOrdersByPage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("erro sqlmock: %v", err)
	}
	defer db.Close()

	repo := &PostgresOrderRepository{db: db}

	now := time.Now().Truncate(time.Second)
	items1 := []model.CartItem{{ProductID: "p1", Name: "n1", Quantity: 2, Price: 10.0}}
	items2 := []model.CartItem{{ProductID: "p2", Name: "n2", Quantity: 1, Price: 15.0}}
	itemsJSON1, _ := json.Marshal(items1)
	itemsJSON2, _ := json.Marshal(items2)

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "items", "total", "status", "created_at", "updated_at",
	}).AddRow(
		"o1", "user1", itemsJSON1, 20.0, model.OrderCreated, now, now,
	).AddRow(
		"o2", "user2", itemsJSON2, 15.0, model.OrderProcessing, now, now,
	)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, user_id, items, total, status, created_at, updated_at")).
		WithArgs(10, 0).
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(25)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM TBLOrders")).
		WillReturnRows(countRows)

	result, err := repo.GetOrdersByPage(context.Background(), 1, 10)
	if err != nil {
		t.Errorf("GetOrdersByPage retornou erro: %v", err)
	}
	if result.TotalCount != 25 {
		t.Errorf("TotalCount = %d, want 25", result.TotalCount)
	}
	if len(result.Orders) != 2 {
		t.Errorf("Orders length = %d, want 2", len(result.Orders))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations não atendidas: %v", err)
	}
}

func TestPostgresOrderRepository_GetOrdersByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("erro sqlmock: %v", err)
	}
	defer db.Close()

	repo := &PostgresOrderRepository{db: db}

	now := time.Now().Truncate(time.Second)
	items := []model.CartItem{{ProductID: "p1", Name: "n1", Quantity: 2, Price: 10.0}}
	itemsJSON, _ := json.Marshal(items)

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "items", "total", "status", "created_at", "updated_at",
	}).AddRow(
		"o1", "user123", itemsJSON, 20.0, model.OrderCreated, now, now,
	)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, user_id, items, total, status, created_at, updated_at")).
		WithArgs("user123", 10, 0).
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM TBLOrders WHERE user_id = $1")).
		WithArgs("user123").
		WillReturnRows(countRows)

	result, err := repo.GetOrdersByUserID(context.Background(), "user123", 1, 10)
	if err != nil {
		t.Errorf("GetOrdersByUserID retornou erro: %v", err)
	}
	if result.TotalCount != 5 {
		t.Errorf("TotalCount = %d, want 5", result.TotalCount)
	}
	if len(result.Orders) != 1 {
		t.Errorf("Orders length = %d, want 1", len(result.Orders))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations não atendidas: %v", err)
	}
}
