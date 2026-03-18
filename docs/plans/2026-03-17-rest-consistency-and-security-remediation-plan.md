# REST Consistency and Security Remediation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Harden security-critical API surfaces and introduce canonical REST route shapes without breaking existing clients.

**Architecture:** Apply non-breaking changes first: keep legacy routes as aliases while introducing canonical REST endpoints, then migrate UI consumers to canonical paths. In parallel, remediate CORS, pagination limits, and timeout middleware safety in backend services.

**Tech Stack:** Go (Gin/Fiber/net/http), React + TypeScript, Docusaurus docs.

### Task 1: Security Remediation Batch (parallel-safe)

**Files:**
- Modify: `services/auth-service/internal/middleware/middleware.go`
- Modify: `services/auth-service/internal/middleware/middleware_test.go`
- Modify: `services/publish-order-service/internal/middleware/cors.go`
- Modify: `services/publish-order-service/internal/middleware/cors_test.go`
- Modify: `services/product-service/main.go`
- Modify: `services/product-service/main_test.go`
- Modify: `services/product-service/internal/handler/product_handler.go`
- Modify: `services/product-service/internal/handler/product_handler_test.go`
- Modify: `services/publish-order-service/internal/middleware/timeout.go`
- Modify: `services/publish-order-service/internal/middleware/timeout_test.go`

**Step 1: Verify failing tests before fix (RED)**
Run:
- `go test ./internal/middleware` in auth-service
- `go test ./internal/middleware` in publish-order-service
- `go test ./internal/handler` in product-service
Expected: failures for insecure/default behavior assumptions.

**Step 2: Implement minimal secure changes**
- Allowlist CORS with `CORS_ALLOWED_ORIGINS` + secure defaults
- Pagination bounds (`page >=1`, `1 <= pageSize <= 100`)
- Replace custom timeout goroutine with standard library safe timeout handling

**Step 3: Verify tests (GREEN)**
Run service-specific tests again and ensure pass.

### Task 2: Canonical Order REST Routes (non-breaking)

**Files:**
- Modify: `services/publish-order-service/main.go`
- Modify: `services/ui-service/src/services/order.service.ts`
- Test: `services/ui-service/src/services/order.service.test.ts`

**Step 1: Add canonical routes while preserving legacy aliases**
Add canonical aliases:
- `POST /api/orders`
- `GET /api/orders`
- `GET /api/me/orders`
- `GET /api/me/orders/{id}`
- `GET /api/me/orders/{id}/events`
- `PATCH /api/orders/{id}/status`
Keep existing `/api/order/*` routes working.

**Step 2: Switch UI API client to canonical routes**
Update `order.service.ts` to consume canonical routes.

**Step 3: Verify tests**
Run:
- `go test ./...` in publish-order-service
- `npm run test -- src/services/order.service.test.ts` in ui-service

### Task 3: REST Audit Report Publication

**Files:**
- Create: `docs-site/docs/04-rest-route-consistency.md`
- Modify: `docs-site/docs/01-overview.md` (Microservices capitalization)

**Step 1: Publish cross-service REST inconsistencies report**
Document:
- Per-service problematic routes
- Proposed canonical replacements
- Migration strategy and deprecation headers
- References to external REST best-practice sources

**Step 2: Verify docs build**
Run: `npm run build` in `docs-site`.

### Task 4: Commit + Push

**Step 1: Commit by concern**
- `fix: harden cors timeout and pagination security controls`
- `refactor: add canonical order routes and align ui clients`
- `docs: add rest route consistency guidance`

**Step 2: Push**
Run `git push origin master`.
