package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/icl00ud/publish-order-service/internal/model"
	"go.uber.org/zap"
)

type OrderRepository interface {
	Save(ctx context.Context, order model.Order) error
	Find(ctx context.Context, id string) (model.Order, error)
	FindByUserID(ctx context.Context, userID, orderID string) (model.Order, error)
	GetOrdersByPage(ctx context.Context, page, pageSize int) (*model.PaginatedOrdersResponse, error)
	GetOrdersByUserID(ctx context.Context, userID string, page, pageSize int) (*model.PaginatedOrdersResponse, error)
	GetOrdersCount(ctx context.Context) (int64, error)
	GetOrdersCountByUserID(ctx context.Context, userID string) (int64, error)
}

type PostgresOrderRepository struct {
	db *sql.DB
}

func (r *PostgresOrderRepository) DB() *sql.DB {
	return r.db
}

func NewOrderRepository(dsn string) (OrderRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(30 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	zap.L().Info("database connection pool configured",
		zap.Int("max_open_conns", 25),
		zap.Int("max_idle_conns", 10))

	return &PostgresOrderRepository{db: db}, nil
}

func (r *PostgresOrderRepository) Save(ctx context.Context, o model.Order) error {
	data, err := json.Marshal(o.Items)
	if err != nil {
		return err
	}
	const q = `
        INSERT INTO TBLOrders(id, user_id, items, total, status, created_at, updated_at)
        VALUES($1,$2,$3,$4,$5,$6,$7)
        ON CONFLICT(id) DO UPDATE
          SET user_id    = EXCLUDED.user_id,
              items      = EXCLUDED.items,
              total      = EXCLUDED.total,
              status     = EXCLUDED.status,
              updated_at = EXCLUDED.updated_at;
    `
	if _, err = r.db.ExecContext(ctx, q,
		o.ID, o.UserID, data, o.Total, o.Status, o.CreatedAt, o.UpdatedAt,
	); err != nil {
		zap.L().Error("order save failed", zap.Error(err))
	}
	return err
}

func (r *PostgresOrderRepository) Find(ctx context.Context, id string) (model.Order, error) {
	var o model.Order
	var data []byte
	const q = `
        SELECT id, user_id, items, total, status, created_at, updated_at
          FROM TBLOrders
         WHERE id = $1;
    `
	row := r.db.QueryRowContext(ctx, q, id)
	if err := row.Scan(&o.ID, &o.UserID, &data, &o.Total, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
		return o, err
	}
	o.Items = []model.CartItem{}
	if len(data) > 0 {
		_ = json.Unmarshal(data, &o.Items)
	}
	return o, nil
}

func (r *PostgresOrderRepository) FindByUserID(ctx context.Context, userID, orderID string) (model.Order, error) {
	var o model.Order
	var data []byte
	const q = `
        SELECT id, user_id, items, total, status, created_at, updated_at
          FROM TBLOrders
         WHERE id = $1 AND user_id = $2;
    `
	row := r.db.QueryRowContext(ctx, q, orderID, userID)
	if err := row.Scan(&o.ID, &o.UserID, &data, &o.Total, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
		return o, err
	}
	o.Items = []model.CartItem{}
	if len(data) > 0 {
		_ = json.Unmarshal(data, &o.Items)
	}
	return o, nil
}

func (r *PostgresOrderRepository) GetOrdersByPage(ctx context.Context, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
	offset := (page - 1) * pageSize

	const q = `
        SELECT id, user_id, items, total, status, created_at, updated_at
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
		if err := rows.Scan(&o.ID, &o.UserID, &data, &o.Total, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
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

func (r *PostgresOrderRepository) GetOrdersByUserID(ctx context.Context, userID string, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
	offset := (page - 1) * pageSize

	const q = `
        SELECT id, user_id, items, total, status, created_at, updated_at
          FROM TBLOrders
         WHERE user_id = $1
         ORDER BY created_at DESC
         LIMIT $2 OFFSET $3;
    `

	rows, err := r.db.QueryContext(ctx, q, userID, pageSize, offset)
	if err != nil {
		zap.L().Error("get orders by user_id failed", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		var data []byte
		if err := rows.Scan(&o.ID, &o.UserID, &data, &o.Total, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
			zap.L().Error("scan order failed", zap.Error(err))
			continue
		}
		o.Items = []model.CartItem{} // Initialize with empty array
		if len(data) > 0 {
			if err := json.Unmarshal(data, &o.Items); err != nil {
				zap.L().Warn("failed to unmarshal items", zap.String("order_id", o.ID), zap.Error(err))
			}
		}
		orders = append(orders, o)
	}

	totalCount, err := r.GetOrdersCountByUserID(ctx, userID)
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

func (r *PostgresOrderRepository) GetOrdersCountByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	const q = `SELECT COUNT(*) FROM TBLOrders WHERE user_id = $1;`

	if err := r.db.QueryRowContext(ctx, q, userID).Scan(&count); err != nil {
		zap.L().Error("get orders count by user_id failed", zap.Error(err))
		return 0, err
	}

	return count, nil
}
