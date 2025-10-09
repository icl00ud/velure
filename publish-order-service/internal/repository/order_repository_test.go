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
