package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/icl00ud/publish-order-service/internal/model"
	"go.uber.org/zap"
)

type OrderRepository interface {
	Save(ctx context.Context, order model.Order) error
	Find(ctx context.Context, id string) (model.Order, error)
	GetOrdersByPage(ctx context.Context, page, pageSize int) (*model.PaginatedOrdersResponse, error)
	GetOrdersCount(ctx context.Context) (int64, error)
}

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(dsn string) (OrderRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresOrderRepository{db: db}, nil
}

func (r *PostgresOrderRepository) Save(ctx context.Context, o model.Order) error {
	data, err := json.Marshal(o.Items)
	if err != nil {
		return err
	}
	const q = `
        INSERT INTO TBLOrders(id, items, total, status, created_at, updated_at)
        VALUES($1,$2,$3,$4,$5,$6)
        ON CONFLICT(id) DO UPDATE
          SET items      = EXCLUDED.items,
              total      = EXCLUDED.total,
              status     = EXCLUDED.status,
              updated_at = EXCLUDED.updated_at;
    `
	if _, err = r.db.ExecContext(ctx, q,
		o.ID, data, o.Total, o.Status, o.CreatedAt, o.UpdatedAt,
	); err != nil {
		zap.L().Error("order save failed", zap.Error(err))
	}
	return err
}

func (r *PostgresOrderRepository) Find(ctx context.Context, id string) (model.Order, error) {
	var o model.Order
	var data []byte
	const q = `
        SELECT id, items, total, status, created_at, updated_at
          FROM TBLOrders
         WHERE id = $1;
    `
	row := r.db.QueryRowContext(ctx, q, id)
	if err := row.Scan(&o.ID, &data, &o.Total, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
		return o, err
	}
	_ = json.Unmarshal(data, &o.Items)
	return o, nil
}

func (r *PostgresOrderRepository) GetOrdersByPage(ctx context.Context, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
	offset := (page - 1) * pageSize

	const q = `
        SELECT id, items, total, status, created_at, updated_at
          FROM TBLOrders
         ORDER BY created_at DESC
         LIMIT $1 OFFSET $2;
    `

	rows, err := r.db.QueryContext(ctx, q, pageSize, offset)
	if err != nil {
		zap.L().Error("get orders by page failed", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		var data []byte
		if err := rows.Scan(&o.ID, &data, &o.Total, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
			zap.L().Error("scan order failed", zap.Error(err))
			continue
		}
		_ = json.Unmarshal(data, &o.Items)
		orders = append(orders, o)
	}

	totalCount, err := r.GetOrdersCount(ctx)
	if err != nil {
		return nil, err
	}

	return model.NewPaginatedOrdersResponse(orders, totalCount, page, pageSize), nil
}

func (r *PostgresOrderRepository) GetOrdersCount(ctx context.Context) (int64, error) {
	var count int64
	const q = `SELECT COUNT(*) FROM TBLOrders;`

	if err := r.db.QueryRowContext(ctx, q).Scan(&count); err != nil {
		zap.L().Error("get orders count failed", zap.Error(err))
		return 0, err
	}

	return count, nil
}
