# process-order-service

Async order processor. Consumes `order.created`, checks inventory, simulates payment, publishes terminal status. No database of its own.

- **Stack:** Go 1.25 · net/http · RabbitMQ
- **Port:** `8081` (health/metrics only)
- **Full docs:** [`docs/microservices/process-order-service.md`](../../docs/microservices/process-order-service.md)

## Flow

```
RabbitMQ "orders" queue
        │
        ▼
  inventory check ──► product-service HTTP
        │
        ▼
  simulated payment
        │
        ▼
  publish status ──► RabbitMQ ──► publish-order-service
```

Statuses: `CREATED` → `PROCESSING` → `COMPLETED` | `FAILED`.

## Local

```bash
go run .                  # requires rabbitmq + product-service reachable
go test ./...
```

Env vars in `.env.example`.
