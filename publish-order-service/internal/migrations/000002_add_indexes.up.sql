CREATE INDEX IF NOT EXISTS idx_orders_created_at ON TBLOrders(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_status ON TBLOrders(status);
CREATE INDEX IF NOT EXISTS idx_orders_status_created_at ON TBLOrders(status, created_at DESC);
