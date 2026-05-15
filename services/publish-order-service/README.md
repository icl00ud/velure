# publish-order-service

Order intake, persistence, event publishing, and real-time status streaming over SSE.

- **Stack:** Go 1.25 · net/http · PostgreSQL (lib/pq, raw SQL) · RabbitMQ · SSE
- **Port:** `8080`
- **Full docs:** [`docs/microservices/publish-order-service.md`](../../docs/microservices/publish-order-service.md)

## Flow

```
POST /api/orders ──► Postgres ──► publish "order.created" to RabbitMQ
                                            │
                                            ▼
                                   process-order-service
                                            │
                                            ▼
                  consume status update ──► Postgres ──► SSE fan-out
```

## Endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/api/orders` | Create order, publish event |
| `GET` | `/api/me/orders` | List user's orders |
| `GET` | `/api/me/orders/{id}` | Order detail |
| `GET` | `/api/me/orders/{id}/events` | SSE status stream |
| `PATCH` | `/api/orders/{id}/status` | Status update (internal) |

## Local

```bash
go run .                  # requires postgres + rabbitmq
go test ./...             # uses testcontainers (see tools.go)
```

Migrations under `migrations/`, run automatically at boot via `internal/database`.
