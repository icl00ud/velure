# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Velure is an e-commerce microservices platform built as a learning project for cloud-native architecture and DevSecOps practices.

**Stack:** Go 1.25+ (Gin/Fiber/net-http) | React 18 + TypeScript + Vite | PostgreSQL, MongoDB, Redis | RabbitMQ | Docker/Kubernetes | Caddy

## Essential Commands

All commands run from repository root.

### Local Development
```bash
make local-up          # Start everything (infra + services + monitoring)
make local-down        # Stop and clean up
```

### AWS Deployment
```bash
make cloud-up          # Deploy full AWS infrastructure (~25 min)
make cloud-down        # Destroy all AWS resources (requires confirmation)
make cloud-urls        # Show access URLs
```

### Service-Specific Commands

**Go Services** (run from each service directory, e.g., `services/auth-service/`):
```bash
go run main.go                           # Start service
go test ./...                            # Run all tests
go test -v -run TestName ./...           # Run single test
go test -race -coverprofile=coverage.out ./...  # With coverage
```

**UI Service** (run from `services/ui-service/`):
```bash
npm run dev              # Start dev server (port 8080)
npm run build            # Production build
npm run lint             # Biome lint
npm run lint:fix         # Auto-fix lint issues
npm run test             # Run tests (vitest)
npm run test -- TestName # Run single test
```

## Architecture

### Microservices & Ports

| Service | Port | Framework | Database | Purpose |
|---------|------|-----------|----------|---------|
| auth-service | 3020 | Go + Gin | PostgreSQL (GORM) + Redis | User auth, JWT, token caching |
| product-service | 3010 | Go + Fiber | MongoDB (native driver) | Product catalog |
| publish-order-service | 8080 | Go + net/http | PostgreSQL (lib/pq, raw SQL) + RabbitMQ | Order creation, SSE status streaming |
| process-order-service | 8081 | Go + net/http | RabbitMQ (no database) | Async order processing, inventory checks |
| ui-service | 80/443 | React + Vite | - | Frontend SPA (nginx in prod) |

### Order Flow (Event-Driven)

This is the core architectural pattern connecting multiple services:

1. Frontend `POST /api/order/create-order` → **publish-order-service**
2. publish-order-service saves order to PostgreSQL, publishes `order.created` event to RabbitMQ (exchange: `orders`)
3. **process-order-service** consumes from RabbitMQ queue `orders`, calls **product-service** HTTP API for inventory checks, processes payment logic
4. process-order-service publishes status update events back to RabbitMQ
5. **publish-order-service** consumes status updates, updates PostgreSQL, broadcasts to frontend via SSE
6. Frontend receives real-time status updates through `GET /api/order/user/order/status?id=X` SSE stream

Order statuses: `CREATED` → `PROCESSING` → `COMPLETED` | `FAILED`

### Shared Code

All Go services use a shared module at `shared/` via Go module replace directive:
```
replace github.com/icl00ud/velure-shared => ../../shared
```
Contains: `logger/` (structured logging with color) and `models/` (shared data models).

### Frontend Architecture

- **Routing:** React Router v6 with `<ProtectedRoute>` wrapper for authenticated pages
- **State:** React Context (AuthContext for auth state) + TanStack React Query v5 (data fetching/caching)
- **Forms:** React Hook Form + Zod schema validation
- **UI Components:** Radix UI + shadcn/ui + Tailwind CSS
- **Linting:** Biome (formatter + linter, configured in `biome.json`)
- **Path alias:** `@` maps to `./src` (configured in vite.config.ts and tsconfig)
- **Key routes:** `/`, `/login`, `/products`, `/products/:category`, `/product/:id`, `/cart`, `/orders`, `/orders/:id`

### Docker & Networking

Local development uses three Docker networks for service isolation:
- `local_auth` — auth-service, postgres, redis, caddy
- `local_order` — order services, product-service, rabbitmq, mongodb, redis
- `local_frontend` — ui-service, caddy

Go services use multi-stage Docker builds (golang:alpine → alpine). UI uses Bun for build, nginx:alpine for serving. Resource limits are set on all containers.

### API Routes (through Caddy reverse proxy)

- Frontend: `https://velure.local`
- Auth API: `https://velure.local/api/auth/*`
- Product API: `https://velure.local/api/product/*`
- Order API: `https://velure.local/api/order/*`

## Critical Notes

### Access Pattern (IMPORTANT)
**Always use:** `https://velure.local` — **Never use:** Direct container URLs

Accessing services directly bypasses Caddy routing and causes 405 errors. Requires `/etc/hosts` entry:
```bash
echo "127.0.0.1 velure.local" | sudo tee -a /etc/hosts
```

### Service Dependencies
- Infrastructure must start before services (handled by `make local-up`)
- RabbitMQ takes 10-15 seconds to initialize
- process-order-service requires product-service for inventory checks via HTTP

### Common Issues
1. **Port conflicts:** Check ports 80, 443, 5432, 27017, 5672, 6379
2. **Self-signed certs:** Accept browser security warning on first access
3. **Direct container access:** Always proxy through Caddy

### Conventions
- **Commits:** Follow [Conventional Commits](https://www.conventionalcommits.org/) (e.g., `feat:`, `fix:`, `refactor:`)
- **Go services:** Clean Architecture layers — `handler/` → `service/` → `repository/`
- **UI formatting:** Biome with double quotes, semicolons always, ES5 trailing commas

## Environment Configuration

- Local config: `infrastructure/local/.env` (copy from `.env.example`)
- Each service reads specific env vars for DB connections, ports, JWT secrets
- RabbitMQ uses separate users per service for isolation (`PUBLISHER_RABBITMQ_USER`, `PROCESS_RABBITMQ_USER`)
