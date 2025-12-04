DROP INDEX IF EXISTS idx_orders_user_id_created_at;
DROP INDEX IF EXISTS idx_orders_user_id;
ALTER TABLE TBLOrders DROP COLUMN IF EXISTS user_id;
