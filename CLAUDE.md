# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Velure is an e-commerce microservices platform built as a learning project to demonstrate modern cloud-native architecture and DevSecOps practices. The system uses Go for backend services, React/TypeScript for the frontend, and runs on Docker locally with Kubernetes deployment to AWS EKS.

**Technology Stack:**
- Backend: Go 1.23+ with Gin/Fiber frameworks
- Frontend: React 18 + TypeScript + Vite + Tailwind CSS
- Databases: PostgreSQL, MongoDB, Redis
- Message Queue: RabbitMQ
- Infrastructure: Docker, Kubernetes, Terraform, AWS EKS
- Reverse Proxy: Caddy with automatic HTTPS

## Common Development Commands

All commands should be run from the repository root unless otherwise specified.

### Quick Start
```bash
make dev                 # Start infrastructure (databases, RabbitMQ, Caddy)
make dev-services        # Run all services with hot reload
make dev-stop            # Stop all containers
make dev-clean           # Clean volumes and data
```

### Development Workflow
```bash
make deps                # Install all dependencies
make build               # Build all services
make test                # Run tests across services
make lint                # Run linters
make format              # Auto-format code
make security            # Run security scans
make health              # Check service health
```

### Service-Specific Commands

**Go Services** (run from service directory):
```bash
go run main.go           # Start service
go test ./...            # Run tests
go mod tidy              # Clean dependencies
gofmt -w .               # Format code
go vet ./...             # Static analysis
```

**UI Service** (run from `services/ui-service/`):
```bash
npm install              # Install dependencies
npm run dev              # Start dev server
npm run build            # Production build
npm run lint             # Biome lint
npm run lint:fix         # Auto-fix issues
npm run format           # Format code
```

### Docker Operations
```bash
make docker-build        # Build all Docker images
docker compose logs -f <service>  # View service logs
```

### Load Testing
```bash
cd tests/load
./run-all-tests.sh       # Run all k6 load tests
k6 run <test-file>.js    # Run specific test
```

## Architecture

### Microservices

**auth-service** (Port 3020)
- User registration, login, JWT generation
- Tech: Go + Gin + PostgreSQL + Redis + GORM
- Entry: `services/auth-service/main.go`

**product-service** (Port 3010)
- Product catalog CRUD, inventory management, caching
- Tech: Go + Fiber + MongoDB + Redis
- Entry: `services/product-service/main.go`

**publish-order-service** (Port 8080)
- Order creation, publishes to RabbitMQ, SSE status updates
- Tech: Go + PostgreSQL + RabbitMQ
- Entry: `services/publish-order-service/main.go`

**process-order-service** (Port 8081)
- Async order processing, consumes from RabbitMQ, payment simulation
- Tech: Go + PostgreSQL + RabbitMQ
- Entry: `services/process-order-service/main.go`

**ui-service** (Port 8080 internal, 80/443 external)
- React SPA with product browsing, cart, checkout
- Tech: React + TypeScript + Vite + TailwindCSS + Radix UI
- Entry: `services/ui-service/src/App.tsx`

### Communication Patterns

1. **Synchronous (HTTP/REST):** Frontend ↔ Backend services, process-order ↔ product-service
2. **Asynchronous (RabbitMQ):** publish-order → process-order (exchange: "orders")
3. **Real-time (SSE):** publish-order → Frontend (order status updates)

All external requests route through Caddy reverse proxy at `https://velure.local`

### Network Architecture
- **Docker Networks:** auth, order, frontend (service isolation)
- **Reverse Proxy:** Caddy handles TLS, routing, CORS, security headers
- **Service Discovery:** Docker Compose DNS resolution

## Code Organization

### Go Services Pattern (Clean Architecture)
```
service-name/
├── main.go                    # Entry point, server setup
├── internal/
│   ├── config/               # Environment configuration
│   ├── models/               # Data models (GORM/MongoDB)
│   ├── handlers/             # HTTP handlers (controllers)
│   ├── services/             # Business logic
│   ├── repository/           # Data access layer
│   ├── middleware/           # Auth, CORS, logging
│   └── database/             # DB connection management
└── migrations/               # SQL migrations (auth-service)
```

**Development Pattern:**
- Follow layered architecture: handlers → services → repository → models
- Keep business logic in services layer
- Use repository pattern for data access
- All handlers should have error handling and validation

### React Service Pattern
```
ui-service/src/
├── pages/                    # Page components (route level)
├── components/               # Reusable UI components
├── services/                 # API clients (fetch calls)
├── hooks/                    # Custom React hooks
├── types/                    # TypeScript type definitions
├── config/                   # App configuration
└── utils/                    # Helper functions
```

## Critical Development Notes

### ⚠️ Access Pattern (IMPORTANT)
**Always use:** `https://velure.local`
**Never use:** Direct container URLs like `ui-service.local.orb.local`

Accessing services directly bypasses Caddy routing and causes 405 Method Not Allowed errors. All API calls must go through the reverse proxy.

### Required /etc/hosts Configuration
```bash
echo "127.0.0.1 velure.local" | sudo tee -a /etc/hosts
```

### Environment Variables
- Configuration in `infrastructure/local/.env` (copy from `.env.example`)
- Each service reads specific env vars for DB connections, ports, JWT secrets
- RabbitMQ uses separate users per service (security isolation)

### API Endpoints
- Frontend: `https://velure.local`
- Auth API: `https://velure.local/api/auth/*`
- Product API: `https://velure.local/api/product/*`
- Order API: `https://velure.local/api/order/*`
- RabbitMQ UI: `http://localhost:15672` (admin/admin_password)

### Service Dependencies
- Infrastructure (databases, RabbitMQ) must start before services
- RabbitMQ takes time to initialize; services retry connections
- process-order-service requires product-service for inventory checks

## Testing

### Unit Tests
- Go: `go test ./...` from service directory
- Coverage enabled in CI/CD with race detection
- React: Tests configured but not yet implemented

### Load Tests
Location: `tests/load/`
- `auth-service-test.js` - Authentication endpoints
- `product-service-test.js` - Product catalog
- `publish-order-service-test.js` - Order creation
- `integrated-load-test.js` - Full user journey
- `run-all-tests.sh` - Execute all tests sequentially

### Integration Tests
Planned in `tests/integration/` (not yet implemented)

## CI/CD Pipeline

**Location:** `.github/workflows/`

### Key Workflows
- **ci-cd.yml** - Main pipeline with path-based service triggers
- **go-service.yml** - Reusable workflow: format, vet, test, security scan, Docker build
- **node-service.yml** - Node.js service workflow
- **security-quality.yml** - Daily security scans (Semgrep, Trivy, gosec, Docker Scout)

### Deployment
- Triggers on push/PR to master
- Multi-platform Docker builds (linux/amd64, linux/arm64)
- Images tagged with: branch, PR number, SHA, latest
- Kubernetes deployment via Helm charts
- AWS EKS deployment via Terraform

## Database Strategy

- **PostgreSQL 17:** Transactional data (auth users, orders)
- **MongoDB 6.0:** Flexible product catalog with nested data
- **Redis 8.0:** Caching layer for frequently accessed products
- Connection pooling configured in each service
- Migrations in `services/auth-service/migrations/`

## Security Practices

- All containers run as non-root users
- JWT-based authentication with refresh tokens
- Separate RabbitMQ users per service (least privilege)
- CORS configured via Caddy
- Security headers (HSTS, CSP, X-Frame-Options)
- Input validation on all endpoints
- Network isolation via Docker networks
- Daily security scans in CI/CD

## Key Documentation Files

- `/README.md` - Main project overview
- `/Makefile` - All available commands
- `/infrastructure/local/README.md` - Local development guide
- `/infrastructure/terraform/README.md` - AWS EKS deployment guide
- `/docs/architecture/ARCHITECTURE.md` - Detailed architecture
- `/docs/DEPLOY_GUIDE.md` - Complete deployment guide

## Common Pitfalls

1. **Port conflicts:** Check ports 80, 443, 5432, 27017, 5672, 6379 are free
2. **Self-signed certs:** Accept browser security warnings on first access
3. **RabbitMQ initialization:** Allow 10-15 seconds for queue setup
4. **Hot reload:** Go services use `air` for auto-reload in dev mode
5. **Direct container access:** Always proxy through Caddy (see Access Pattern above)
