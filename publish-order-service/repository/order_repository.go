package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/icl00ud/publish-order-service/pkg/model"
	"go.uber.org/zap"
)

type OrderRepository interface {
	Save(ctx context.Context, order model.Order) error
	Find(ctx context.Context, id string) (model.Order, error)
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
