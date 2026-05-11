# local-dev Hot-Reload Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `make local-dev` that starts the full local stack with Air hot-reload for all Go services and Vite HMR for the UI, without rebuilding containers on source changes.

**Architecture:** A `docker-compose.dev.yaml` overlay replaces the 5 application service definitions (auth, product, publish-order, process-order, ui) while leaving infra containers (postgres, redis, rabbitmq, mongodb, caddy) untouched. Go services use a shared dev image with Air installed; UI uses `oven/bun` running `bun run dev`. Two lightweight dev Dockerfiles live in `infrastructure/local/`.

**Tech Stack:** Air (`github.com/air-verse/air`) for Go hot-reload, Vite 5 HMR for UI, Docker Compose override merge, `golang:1.25.5-alpine`, `oven/bun:1-alpine`.

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `services/auth-service/.air.toml` | Create | Air watcher config for auth-service |
| `services/publish-order-service/.air.toml` | Create | Air watcher config for publish-order-service |
| `services/process-order-service/.air.toml` | Create | Air watcher config for process-order-service |
| `infrastructure/local/Dockerfile.go.dev` | Create | Shared dev image: Go + Air |
| `infrastructure/local/Dockerfile.ui.dev` | Create | Dev image: Bun (no nginx, no build step) |
| `infrastructure/local/docker-compose.dev.yaml` | Create | Compose overlay: override 5 app services for dev |
| `services/ui-service/vite.config.ts` | Modify | Add `hmr.clientPort` so HMR WebSocket routes through Caddy on port 80 |
| `Makefile` | Modify | Add `local-dev` target + update `.PHONY` |

---

### Task 1: Air configs for three Go services

**Files:**
- Create: `services/auth-service/.air.toml`
- Create: `services/publish-order-service/.air.toml`
- Create: `services/process-order-service/.air.toml`

- [ ] **Step 1: Create `services/auth-service/.air.toml`**

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ."
  delay = 0
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

- [ ] **Step 2: Create `services/publish-order-service/.air.toml`**

Identical content to Step 1 — copy the same file verbatim.

- [ ] **Step 3: Create `services/process-order-service/.air.toml`**

Identical content to Step 1 — copy the same file verbatim.

- [ ] **Step 4: Commit**

```bash
git add services/auth-service/.air.toml services/publish-order-service/.air.toml services/process-order-service/.air.toml
git commit -m "chore: add Air hot-reload config to Go services"
```

---

### Task 2: Dev Dockerfiles

**Files:**
- Create: `infrastructure/local/Dockerfile.go.dev`
- Create: `infrastructure/local/Dockerfile.ui.dev`

- [ ] **Step 1: Create `infrastructure/local/Dockerfile.go.dev`**

This image is shared by all four Go services. It installs Air once; the container mounts source at startup.

```dockerfile
FROM golang:1.25.5-alpine
RUN apk add --no-cache git && go install github.com/air-verse/air@latest
WORKDIR /workspace
```

- [ ] **Step 2: Create `infrastructure/local/Dockerfile.ui.dev`**

Minimal Bun image. Source code is mounted at runtime; no build step.

```dockerfile
FROM oven/bun:1-alpine
WORKDIR /app
```

- [ ] **Step 3: Commit**

```bash
git add infrastructure/local/Dockerfile.go.dev infrastructure/local/Dockerfile.ui.dev
git commit -m "chore: add dev Dockerfiles for Go and UI services"
```

---

### Task 3: docker-compose.dev.yaml overlay

**Files:**
- Create: `infrastructure/local/docker-compose.dev.yaml`

- [ ] **Step 1: Create `infrastructure/local/docker-compose.dev.yaml`**

The overlay only defines the 5 app services. Docker Compose merges this with `docker-compose.yaml`: environment variables, networks, ports, and `depends_on` from the base file are preserved; `build`, `command`, `volumes`, and `working_dir` are replaced by the values below.

```yaml
services:
  # -------------------------
  # Go services — Air hot-reload
  # Repo root mounted at /workspace so go.mod replace directive
  # "../../shared" resolves to /workspace/shared without changes.
  # -------------------------
  auth-service:
    build:
      context: ../..
      dockerfile: infrastructure/local/Dockerfile.go.dev
    working_dir: /workspace/services/auth-service
    command: air
    volumes:
      - ../..:/workspace
    restart: unless-stopped

  product-service:
    build:
      context: ../..
      dockerfile: infrastructure/local/Dockerfile.go.dev
    working_dir: /workspace/services/product-service
    command: air
    volumes:
      - ../..:/workspace
    restart: unless-stopped

  publish-order-service:
    build:
      context: ../..
      dockerfile: infrastructure/local/Dockerfile.go.dev
    working_dir: /workspace/services/publish-order-service
    command: air
    volumes:
      - ../..:/workspace
    restart: unless-stopped

  process-order-service:
    build:
      context: ../..
      dockerfile: infrastructure/local/Dockerfile.go.dev
    working_dir: /workspace/services/process-order-service
    command: air
    volumes:
      - ../..:/workspace
    restart: unless-stopped

  # -------------------------
  # UI — Vite dev server (HMR)
  # node_modules in named volume to avoid host/container arch mismatch.
  # -------------------------
  ui-service:
    build:
      context: ../..
      dockerfile: infrastructure/local/Dockerfile.ui.dev
    working_dir: /app
    command: sh -c "bun install && bun run dev"
    volumes:
      - ../../services/ui-service:/app
      - ui_node_modules:/app/node_modules
    environment:
      VITE_PRODUCT_SERVICE_URL: /api/products
      VITE_AUTHENTICATION_SERVICE_URL: /api
      VITE_ORDER_SERVICE_URL: /api/orders
      VITE_HMR_CLIENT_PORT: "80"
    restart: unless-stopped

volumes:
  ui_node_modules:
```

- [ ] **Step 2: Commit**

```bash
git add infrastructure/local/docker-compose.dev.yaml
git commit -m "chore: add docker-compose.dev.yaml overlay for hot-reload stack"
```

---

### Task 4: Vite HMR client port fix

**Files:**
- Modify: `services/ui-service/vite.config.ts`

**Why this is needed:** Vite embeds a WebSocket URL in the browser bundle for HMR. Without this fix, Vite tells the browser to connect to `ws://localhost:8080` (Vite's internal port), which the browser can't reach because Caddy sits in front on port 80. Setting `clientPort: 80` tells Vite to instruct the browser to open the HMR WebSocket at port 80 instead, which Caddy then proxies to Vite on port 8080.

When `VITE_HMR_CLIENT_PORT` is not set (e.g., running `npm run dev` locally without Caddy), `parseInt` falls back to `8080` — existing local dev behavior is unchanged.

- [ ] **Step 1: Read the current server config block in `services/ui-service/vite.config.ts`**

Current block (lines 7–11):
```ts
  server: {
    host: "::",
    port: 8080,
  },
```

- [ ] **Step 2: Replace the server config block**

```ts
  server: {
    host: "::",
    port: 8080,
    hmr: {
      clientPort: parseInt(process.env.VITE_HMR_CLIENT_PORT || "8080"),
    },
  },
```

- [ ] **Step 3: Verify no TypeScript errors**

Run from `services/ui-service/`:
```bash
npm run build 2>&1 | head -20
```
Expected: build completes without type errors on vite.config.ts.

- [ ] **Step 4: Commit**

```bash
git add services/ui-service/vite.config.ts
git commit -m "fix: configure Vite HMR clientPort via env var for Caddy proxy compatibility"
```

---

### Task 5: Makefile local-dev target

**Files:**
- Modify: `Makefile`

- [ ] **Step 1: Add `local-dev` to the `.PHONY` line**

Current line (line 8):
```makefile
.PHONY: help local-up local-down cloud-up cloud-down cloud-urls docs-up docs-down test docs-build lint check
```

Replace with:
```makefile
.PHONY: help local-up local-down local-dev cloud-up cloud-down cloud-urls docs-up docs-down test docs-build lint check
```

- [ ] **Step 2: Add `local-dev` target after `local-down`**

Add after the `local-down` target block (after line 74), before the `# CLOUD` section comment:

```makefile
local-dev: ## Start stack with hot-reload (Air for Go services, Vite HMR for UI)
	@echo "Starting local dev environment with hot-reload..."
	@echo ""
	@echo "Creating Docker networks..."
	@docker network create local_auth 2>/dev/null || echo "  local_auth already exists"
	@docker network create local_order 2>/dev/null || echo "  local_order already exists"
	@docker network create local_frontend 2>/dev/null || echo "  local_frontend already exists"
	@echo ""
	@echo "Building dev images and starting services..."
	@echo "Note: First run builds dev images (~1-2 min). Subsequent starts are faster."
	cd infrastructure/local && docker-compose \
		-f docker-compose.yaml \
		-f docker-compose.dev.yaml \
		up --build -d
	@echo ""
	@echo "Dev environment is ready."
	@echo ""
	@echo "Access URLs:"
	@echo ""
	@echo "Application:  http://localhost"
	@echo "RabbitMQ:     http://localhost:15672 (admin/admin_password)"
	@echo ""
	@echo "Hot-reload:"
	@echo "  Go services  — Air rebuilds on .go file change (~1s)"
	@echo "  UI           — Vite HMR updates browser instantly on .tsx/.ts change"
	@echo ""
	@echo "Watch logs:  docker logs -f <container-name>"
	@echo "Stop:        make local-down"
	@echo ""
```

- [ ] **Step 3: Commit**

```bash
git add Makefile
git commit -m "feat: add make local-dev with Air and Vite hot-reload"
```

---

### Task 6: Smoke test

No code changes. Verify the stack starts and hot-reload works end-to-end.

- [ ] **Step 1: Start the dev stack**

```bash
make local-dev
```

Expected: all containers start (first run may take 1-2 min while dev images build). No fatal errors in output.

- [ ] **Step 2: Check all containers are running**

```bash
docker ps --format "table {{.Names}}\t{{.Status}}" | grep -E "(auth|product|publish|process|ui-service)"
```

Expected: 5 rows, all `Up`.

- [ ] **Step 3: Verify the app loads**

```bash
curl -s -o /dev/null -w "%{http_code}" http://localhost/
```

Expected: `200`

- [ ] **Step 4: Verify Go hot-reload (auth-service)**

Add a harmless log line to `services/auth-service/main.go` — e.g., add a blank comment anywhere — then save.

```bash
docker logs -f auth-service 2>&1 | head -20
```

Expected within ~2s: Air prints `building...` then `running...` with the new binary.

Remove the dummy change and save again to confirm second rebuild.

- [ ] **Step 5: Verify Vite HMR**

Open `http://localhost` in a browser. Open DevTools → Network → WS tab. Verify a WebSocket connection to `ws://localhost/` is open (this is Vite's HMR socket through Caddy).

Make a visible change to `services/ui-service/src/pages/Index.tsx` — e.g., change any text string. Save the file.

Expected: browser updates the text without a full page reload (HMR).

- [ ] **Step 6: Stop the stack**

```bash
make local-down
```

Expected: all containers stop and volumes are removed.
