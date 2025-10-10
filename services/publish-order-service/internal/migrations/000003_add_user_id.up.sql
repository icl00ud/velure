ALTER TABLE TBLOrders ADD COLUMN user_id VARCHAR(255);
CREATE INDEX idx_orders_user_id ON TBLOrders(user_id);
