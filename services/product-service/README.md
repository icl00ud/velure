# product-service

Product catalog and inventory lookup. Read-heavy, document-oriented.

- **Stack:** Go 1.25 · Fiber · MongoDB · Redis
- **Port:** `3010`
- **Full docs:** [`docs/microservices/product-service.md`](../../docs/microservices/product-service.md)

## Endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/products` | List / filter products |
| `GET` | `/api/products/:id` | Product detail |
| `POST` | `/api/products/:id/reserve` | Inventory decrement (called by process-order) |

## Local

```bash
go run .                  # requires mongodb (see infrastructure/local)
go test ./...
```

Seed data: `mongo-init.js` runs automatically on first Mongo container start. Env vars in `.env.example`.
