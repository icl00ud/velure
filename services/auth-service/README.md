# auth-service

Authentication and session management. Issues JWTs, validates tokens, owns the `users` table.

- **Stack:** Go 1.25 · Gin · GORM · PostgreSQL · Redis (JWT cache)
- **Port:** `3020`
- **Full docs:** [`docs/microservices/auth-service.md`](../../docs/microservices/auth-service.md)

## Endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/api/users` | Register |
| `POST` | `/api/sessions` | Login → JWT |
| `GET` | `/api/sessions/validate` | Validate token (cached in Redis) |

## Local

```bash
go run .                  # requires postgres + redis (see infrastructure/local)
go test ./...
make migrate-up           # apply SQL migrations
make load-test            # k6 smoke test
```

Env vars in `.env.example`. Migrations under `migrations/`.
