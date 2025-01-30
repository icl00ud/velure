CREATE TABLE IF NOT EXISTS TBLOrders (
    id VARCHAR(255) PRIMARY KEY,
    items JSONB NOT NULL,
    total INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);