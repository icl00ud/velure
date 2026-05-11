# local-dev Hot-Reload Design

**Date:** 2026-05-11
**Status:** Approved

## Goal

Add a `make local-dev` target that starts the full local stack with hot-reload for all application services, eliminating the need to rebuild and restart containers on source code changes.

## Approach

docker-compose overlay (`docker-compose.dev.yaml`) that replaces the 5 application service definitions while reusing infra services (postgres, redis, rabbitmq, mongodb, caddy) unchanged.

- **Go services**: Air watcher (`github.com/air-verse/air`) detects `.go` file changes, recompiles, and restarts the binary in ~1s.
- **UI service**: Vite dev server (`bun run dev`) with native HMR replaces the nginx+static-build container.

Caddy config is unchanged — Vite already listens on port 8080, matching the existing `ui-service:8080` upstream.

## Files

### New files

| File | Purpose |
|------|---------|
| `infrastructure/local/docker-compose.dev.yaml` | Compose overlay for dev services |
| `services/auth-service/.air.toml` | Air config (copy of product-service pattern) |
| `services/publish-order-service/.air.toml` | Air config |
| `services/process-order-service/.air.toml` | Air config |

### Modified files

| File | Change |
|------|--------|
| `Makefile` | Add `local-dev` target |

`services/product-service/.air.toml` already exists — no change needed.

## docker-compose.dev.yaml

Overrides only application services. Each Go service:

- Image: `golang:1.25.5-alpine` with Air pre-installed via `go install`
- Volume: repo root mounted at `/workspace` so `../../shared` in go.mod resolves to `/workspace/shared` without any go.mod changes
- Working dir: `/workspace/services/<service-name>`
- Command: `air` (reads `.air.toml` from working dir)
- All environment variables inherited from base compose via the overlay merge

UI service:

- Image: `oven/bun:1-alpine`
- Volume: `services/ui-service` mounted at `/app`
- Command: `bun run dev`
- Port 8080 exposed — Caddy proxy unchanged

## Air config (.air.toml)

All services use the same pattern as `product-service/.air.toml`:

```toml
[build]
  cmd = "go build -o ./tmp/main ."
  bin = "./tmp/main"
  include_ext = ["go", "tpl", "tmpl", "html"]
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_regex = ["_test.go"]
```

## Makefile target

```makefile
local-dev: ## Start stack with hot-reload (Air for Go, Vite HMR for UI)
	@docker network create local_auth 2>/dev/null || true
	@docker network create local_order 2>/dev/null || true
	@docker network create local_frontend 2>/dev/null || true
	cd infrastructure/local && docker-compose \
	    -f docker-compose.yaml \
	    -f docker-compose.dev.yaml \
	    up --build -d
```

`--build` ensures dev images are rebuilt if Dockerfiles change. After first build, Air and Vite watch for file changes without needing image rebuilds.

## Data Flow

```
Source file change
       │
       ▼
  [Go service]          [UI service]
  Air detects .go       Vite HMR detects .ts/.tsx
  go build (~1s)        Hot module replace (<100ms)
  restart binary        Browser updates without reload
       │                      │
       ▼                      ▼
  Same port              Same port 8080
  Same networks          Caddy proxy unchanged
```

## Constraints

- First `make local-dev` builds dev images; subsequent starts reuse the image layer cache.
- Go module cache is stored inside the container — lost on `docker-compose down`. Use named volume for `GOPATH` cache to speed up rebuilds if needed (out of scope for this iteration).
- Vite dev server does not serve a `/health` endpoint. Caddy's health check for `ui-service` will log failures but still proxies requests (single upstream behavior).
- `shared/` module changes trigger Air rebuild only on services that import it, provided their `.air.toml` watches the correct paths. Since we mount the full repo root, Air can watch `../../shared` relative to the service dir.
