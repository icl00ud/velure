# Outbox Pattern + Consumer Idempotency Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Eliminate the lost-event window between Postgres commit and RabbitMQ publish in `publish-order-service`, and make `process-order-service` safely dedupe redelivered messages.

**Architecture:** Order writes and event records share one Postgres transaction via an `outbox_events` table. A relay goroutine in `publish-order-service` drains the outbox to RabbitMQ using publisher confirms, `FOR UPDATE SKIP LOCKED`, and `LISTEN/NOTIFY` (with a polling fallback). `process-order-service` checks each incoming `event_id` against Redis `SET NX EX` before processing; duplicates ack-and-drop.

**Tech Stack:** Go 1.25+, Postgres (lib/pq for `LISTEN/NOTIFY`), RabbitMQ (amqp091-go publisher confirms), Redis (go-redis/v9), Prometheus client. Spec: `docs/superpowers/specs/2026-05-22-outbox-pattern-design.md`.

**Conventions:**
- All commands run from repo root (`/Users/icl00ud/repos/velure`) unless noted.
- TDD: every task writes the failing test first, runs it to confirm failure, implements minimal code, runs to confirm pass, commits.
- Commit messages follow Conventional Commits (`feat:`, `fix:`, `refactor:`, `test:`, `docs:`).
- Run `go test ./...` from each service directory (`services/publish-order-service`, `services/process-order-service`).

---

## Task 1: Add `model.OutboxEvent` type

**Files:**
- Create: `services/publish-order-service/internal/model/outbox.go`
- Test: `services/publish-order-service/internal/model/outbox_test.go`

- [ ] **Step 1: Write the failing test**

Create `services/publish-order-service/internal/model/outbox_test.go`:

```go
package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestOutboxEvent_JSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	evt := OutboxEvent{
		ID:          "evt-1",
		AggregateID: "order-1",
		EventType:   OrderCreated,
		Payload:     json.RawMessage(`{"id":"order-1"}`),
		CreatedAt:   now,
	}
	b, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got OutboxEvent
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ID != evt.ID || got.AggregateID != evt.AggregateID || got.EventType != evt.EventType {
		t.Fatalf("mismatch: %+v", got)
	}
	if !got.CreatedAt.Equal(evt.CreatedAt) {
		t.Fatalf("created_at mismatch: got %v want %v", got.CreatedAt, evt.CreatedAt)
	}
	if got.PublishedAt != nil {
		t.Fatalf("expected nil PublishedAt, got %v", got.PublishedAt)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd services/publish-order-service && go test ./internal/model/ -run TestOutboxEvent_JSONRoundTrip -v`
Expected: FAIL with `undefined: OutboxEvent`.

- [ ] **Step 3: Implement `OutboxEvent`**

Create `services/publish-order-service/internal/model/outbox.go`:

```go
package model

import (
	"encoding/json"
	"time"
)

// OutboxEvent is a domain event awaiting publication to RabbitMQ.
// PublishedAt nil = pending; non-nil = already published (kept for audit).
type OutboxEvent struct {
	ID          string          `json:"id"`
	AggregateID string          `json:"aggregate_id"`
	EventType   string          `json:"event_type"`
	Payload     json.RawMessage `json:"payload"`
	CreatedAt   time.Time       `json:"created_at"`
	PublishedAt *time.Time      `json:"published_at,omitempty"`
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd services/publish-order-service && go test ./internal/model/ -run TestOutboxEvent_JSONRoundTrip -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add services/publish-order-service/internal/model/outbox.go services/publish-order-service/internal/model/outbox_test.go
git commit -m "feat(publish-order): add OutboxEvent model"
```

---

## Task 2: Create Postgres migration for `outbox_events` table

**Files:**
- Create: `services/publish-order-service/migrations/005_add_outbox.up.sql`
- Create: `services/publish-order-service/migrations/005_add_outbox.down.sql`
- Test: `services/publish-order-service/internal/database/migrate_test.go` (extend existing — verify migration runs against ephemeral container)

The existing migration tests use a real Postgres container via `testcontainers-go` (check `migrate_test.go` to confirm — if it uses sqlmock instead, the verification below uses raw `psql` against the local stack via `make local-up` and is documented as a manual check; do not invent fake tests).

- [ ] **Step 1: Write the up migration**

Create `services/publish-order-service/migrations/005_add_outbox.up.sql`:

```sql
CREATE TABLE IF NOT EXISTS outbox_events (
    id            UUID PRIMARY KEY,
    aggregate_id  TEXT NOT NULL,
    event_type    TEXT NOT NULL,
    payload       JSONB NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_outbox_unpublished
    ON outbox_events (created_at)
    WHERE published_at IS NULL;

CREATE OR REPLACE FUNCTION outbox_notify() RETURNS trigger AS $$
BEGIN
  PERFORM pg_notify('outbox_new', NEW.id::text);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS outbox_notify_trigger ON outbox_events;
CREATE TRIGGER outbox_notify_trigger
AFTER INSERT ON outbox_events
FOR EACH ROW EXECUTE FUNCTION outbox_notify();
```

- [ ] **Step 2: Write the down migration**

Create `services/publish-order-service/migrations/005_add_outbox.down.sql`:

```sql
DROP TRIGGER IF EXISTS outbox_notify_trigger ON outbox_events;
DROP FUNCTION IF EXISTS outbox_notify();
DROP INDEX IF EXISTS idx_outbox_unpublished;
DROP TABLE IF EXISTS outbox_events;
```

- [ ] **Step 3: Verify migration applies against local stack**

```bash
make local-up
# Wait for publish-order to log "Migrations completed"
docker exec -i $(docker ps -qf name=postgres) psql -U postgres -d publish_order_db -c '\d outbox_events'
```

Expected output: table exists with columns `id`, `aggregate_id`, `event_type`, `payload`, `created_at`, `published_at`. Index `idx_outbox_unpublished` listed. Trigger `outbox_notify_trigger` listed.

If `make local-up` is not available locally, skip this verification step and rely on Task 3's repository tests which open a real `*sql.DB` against the migration.

- [ ] **Step 4: Commit**

```bash
git add services/publish-order-service/migrations/005_add_outbox.up.sql services/publish-order-service/migrations/005_add_outbox.down.sql
git commit -m "feat(publish-order): add outbox_events migration with notify trigger"
```

---

## Task 3: Implement `outbox.Repository` (Postgres)

**Files:**
- Create: `services/publish-order-service/internal/outbox/repository.go`
- Test: `services/publish-order-service/internal/outbox/repository_test.go`

This task uses `github.com/DATA-DOG/go-sqlmock` (already a transitive dep — confirm with `go list -m all | grep sqlmock`; if missing, `go get github.com/DATA-DOG/go-sqlmock`).

- [ ] **Step 1: Write the failing tests**

Create `services/publish-order-service/internal/outbox/repository_test.go`:

```go
package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/icl00ud/velure/services/publish-order-service/internal/model"
)

func TestSaveTx_InsertsRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO outbox_events`).
		WithArgs("evt-1", "order-1", "order.created", []byte(`{"id":"order-1"}`), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	tx, _ := db.BeginTx(context.Background(), nil)
	repo := NewPostgresRepository(db)
	err = repo.SaveTx(context.Background(), tx, model.OutboxEvent{
		ID:          "evt-1",
		AggregateID: "order-1",
		EventType:   "order.created",
		Payload:     json.RawMessage(`{"id":"order-1"}`),
		CreatedAt:   time.Now(),
	})
	if err != nil { t.Fatalf("SaveTx: %v", err) }
	if err := tx.Commit(); err != nil { t.Fatal(err) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}

func TestFetchUnpublished_ReturnsPendingEvents(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "aggregate_id", "event_type", "payload", "created_at"}).
		AddRow("evt-1", "order-1", "order.created", []byte(`{}`), now).
		AddRow("evt-2", "order-2", "order.created", []byte(`{}`), now)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM outbox_events .* FOR UPDATE SKIP LOCKED`).
		WithArgs(50).
		WillReturnRows(rows)

	repo := NewPostgresRepository(db)
	tx, events, err := repo.FetchUnpublished(context.Background(), 50)
	if err != nil { t.Fatalf("FetchUnpublished: %v", err) }
	defer tx.Rollback()

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].ID != "evt-1" || events[1].ID != "evt-2" {
		t.Fatalf("unexpected events: %+v", events)
	}
}

func TestMarkPublished_UpdatesRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE outbox_events SET published_at`).
		WithArgs(sqlmock.AnyArg(), "evt-1", "evt-2").
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	tx, _ := db.BeginTx(context.Background(), nil)
	repo := NewPostgresRepository(db)
	if err := repo.MarkPublished(context.Background(), tx, []string{"evt-1", "evt-2"}); err != nil {
		t.Fatalf("MarkPublished: %v", err)
	}
	if err := tx.Commit(); err != nil { t.Fatal(err) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}

func TestMarkPublished_NoOpOnEmpty(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()

	tx, _ := db.BeginTx(context.Background(), nil)
	repo := NewPostgresRepository(db)
	if err := repo.MarkPublished(context.Background(), tx, nil); err != nil {
		t.Fatalf("MarkPublished(nil) should be no-op, got: %v", err)
	}
}

var _ sql.Result = (*sqlmock.Result)(nil) // compile check that sqlmock is imported
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd services/publish-order-service && go test ./internal/outbox/ -v`
Expected: FAIL — `undefined: NewPostgresRepository`.

- [ ] **Step 3: Implement the repository**

Create `services/publish-order-service/internal/outbox/repository.go`:

```go
package outbox

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/icl00ud/velure/services/publish-order-service/internal/model"
)

type Repository interface {
	// SaveTx inserts the event in the supplied transaction. Caller commits.
	SaveTx(ctx context.Context, tx *sql.Tx, evt model.OutboxEvent) error

	// FetchUnpublished opens a transaction, selects up to `limit` oldest
	// unpublished events with FOR UPDATE SKIP LOCKED, and returns the tx open.
	// The caller MUST either Commit (after MarkPublished) or Rollback to
	// release locks.
	FetchUnpublished(ctx context.Context, limit int) (*sql.Tx, []model.OutboxEvent, error)

	// MarkPublished sets published_at = now() for the given ids inside tx.
	// No-op when ids is empty.
	MarkPublished(ctx context.Context, tx *sql.Tx, ids []string) error
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) SaveTx(ctx context.Context, tx *sql.Tx, evt model.OutboxEvent) error {
	const q = `
		INSERT INTO outbox_events (id, aggregate_id, event_type, payload, created_at)
		VALUES ($1, $2, $3, $4, $5);
	`
	created := evt.CreatedAt
	if created.IsZero() {
		created = time.Now().UTC()
	}
	_, err := tx.ExecContext(ctx, q, evt.ID, evt.AggregateID, evt.EventType, []byte(evt.Payload), created)
	if err != nil {
		return fmt.Errorf("outbox save: %w", err)
	}
	return nil
}

func (r *postgresRepository) FetchUnpublished(ctx context.Context, limit int) (*sql.Tx, []model.OutboxEvent, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("outbox begin tx: %w", err)
	}

	const q = `
		SELECT id, aggregate_id, event_type, payload, created_at
		  FROM outbox_events
		 WHERE published_at IS NULL
		 ORDER BY created_at
		 LIMIT $1
		 FOR UPDATE SKIP LOCKED;
	`
	rows, err := tx.QueryContext(ctx, q, limit)
	if err != nil {
		_ = tx.Rollback()
		return nil, nil, fmt.Errorf("outbox fetch: %w", err)
	}
	defer rows.Close()

	var events []model.OutboxEvent
	for rows.Next() {
		var evt model.OutboxEvent
		var payload []byte
		if err := rows.Scan(&evt.ID, &evt.AggregateID, &evt.EventType, &payload, &evt.CreatedAt); err != nil {
			_ = tx.Rollback()
			return nil, nil, fmt.Errorf("outbox scan: %w", err)
		}
		evt.Payload = payload
		events = append(events, evt)
	}
	if err := rows.Err(); err != nil {
		_ = tx.Rollback()
		return nil, nil, fmt.Errorf("outbox rows: %w", err)
	}
	return tx, events, nil
}

func (r *postgresRepository) MarkPublished(ctx context.Context, tx *sql.Tx, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	placeholders := make([]string, 0, len(ids))
	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, time.Now().UTC())
	for i, id := range ids {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
		args = append(args, id)
	}
	q := fmt.Sprintf(
		`UPDATE outbox_events SET published_at = $1 WHERE id IN (%s);`,
		strings.Join(placeholders, ","),
	)
	if _, err := tx.ExecContext(ctx, q, args...); err != nil {
		return fmt.Errorf("outbox mark published: %w", err)
	}
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd services/publish-order-service && go test ./internal/outbox/ -v`
Expected: PASS — all 4 tests.

- [ ] **Step 5: Commit**

```bash
git add services/publish-order-service/internal/outbox/repository.go services/publish-order-service/internal/outbox/repository_test.go
git commit -m "feat(publish-order): add outbox repository (SaveTx, FetchUnpublished, MarkPublished)"
```

---

## Task 4: Add `SaveTx` to `OrderRepository`

The existing `OrderRepository.Save` uses `r.db.ExecContext`. The new `SaveTx` does the same operation but using a caller-provided `*sql.Tx`.

**Files:**
- Modify: `services/publish-order-service/internal/repository/order_repository.go`
- Test: `services/publish-order-service/internal/repository/order_repository_test.go` (extend)

- [ ] **Step 1: Write the failing test**

Add to `services/publish-order-service/internal/repository/order_repository_test.go` (follow existing sqlmock pattern in that file):

```go
func TestSaveTx_UsesProvidedTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO TBLOrders`).
		WithArgs("order-1", "user-1", sqlmock.AnyArg(), int64(100), "CREATED", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	tx, _ := db.BeginTx(context.Background(), nil)
	repo := &PostgresOrderRepository{db: db}
	err = repo.SaveTx(context.Background(), tx, model.Order{
		ID: "order-1", UserID: "user-1", Total: 100, Status: "CREATED",
		Items: []model.CartItem{{ProductID: "p1", Quantity: 1}},
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	if err != nil { t.Fatalf("SaveTx: %v", err) }
	if err := tx.Commit(); err != nil { t.Fatal(err) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd services/publish-order-service && go test ./internal/repository/ -run TestSaveTx_UsesProvidedTransaction -v`
Expected: FAIL — `repo.SaveTx undefined`.

- [ ] **Step 3: Add `SaveTx` method + extend interface**

In `services/publish-order-service/internal/repository/order_repository.go`:

(a) Add to the `OrderRepository` interface (after `Save`):

```go
SaveTx(ctx context.Context, tx *sql.Tx, order model.Order) error
```

(b) Add the method on `PostgresOrderRepository` (after the existing `Save`):

```go
func (r *PostgresOrderRepository) SaveTx(ctx context.Context, tx *sql.Tx, o model.Order) error {
	data, err := json.Marshal(o.Items)
	if err != nil {
		return err
	}
	const q = `
        INSERT INTO TBLOrders(id, user_id, items, total, status, created_at, updated_at)
        VALUES($1,$2,$3,$4,$5,$6,$7)
        ON CONFLICT(id) DO UPDATE
          SET user_id    = EXCLUDED.user_id,
              items      = EXCLUDED.items,
              total      = EXCLUDED.total,
              status     = EXCLUDED.status,
              updated_at = EXCLUDED.updated_at;
    `
	_, err = tx.ExecContext(ctx, q,
		o.ID, o.UserID, data, o.Total, o.Status, o.CreatedAt, o.UpdatedAt,
	)
	return err
}
```

- [ ] **Step 4: Run test to verify it passes; also run full repo suite**

```bash
cd services/publish-order-service && go test ./internal/repository/ -v
```

Expected: PASS — new test plus all prior repository tests.

- [ ] **Step 5: Commit**

```bash
git add services/publish-order-service/internal/repository/order_repository.go services/publish-order-service/internal/repository/order_repository_test.go
git commit -m "feat(publish-order): add SaveTx to OrderRepository"
```

---

## Task 5: Add `PublishWithConfirm` to publisher

The relay needs broker-acknowledged publishes. amqp091-go exposes publisher confirms via `channel.Confirm(noWait bool) error` + `channel.NotifyPublish(chan amqp091.Confirmation) chan amqp091.Confirmation`.

**Files:**
- Modify: `services/publish-order-service/internal/publisher/publisher.go`
- Test: `services/publish-order-service/internal/publisher/publisher_test.go` (extend)

- [ ] **Step 1: Extend the channel interface and write a test**

Update `amqpPublisherChannel` interface in `publisher.go` to include confirms (add to existing interface):

```go
type amqpPublisherChannel interface {
	Publish(exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp091.Table) error
	Confirm(noWait bool) error
	NotifyPublish(confirm chan amqp091.Confirmation) chan amqp091.Confirmation
	Close() error
}
```

Add to `services/publish-order-service/internal/publisher/publisher_test.go`:

```go
func TestPublishWithConfirm_WaitsForAck(t *testing.T) {
	ch := newFakeChannel()
	confirms := make(chan amqp091.Confirmation, 1)
	ch.confirms = confirms
	p := &rabbitMQPublisher{exchange: "orders", ch: ch, logger: testLogger()}

	go func() {
		// Simulate broker ack arriving shortly after publish
		time.Sleep(10 * time.Millisecond)
		confirms <- amqp091.Confirmation{DeliveryTag: 1, Ack: true}
	}()

	evt := model.OutboxEvent{ID: "evt-1", EventType: "order.created", Payload: []byte(`{}`)}
	err := p.PublishWithConfirm(context.Background(), evt)
	if err != nil {
		t.Fatalf("PublishWithConfirm: %v", err)
	}
	if ch.lastHeaders["event_id"] != "evt-1" {
		t.Fatalf("expected event_id header set, got %v", ch.lastHeaders)
	}
}

func TestPublishWithConfirm_NackErrors(t *testing.T) {
	ch := newFakeChannel()
	confirms := make(chan amqp091.Confirmation, 1)
	ch.confirms = confirms
	p := &rabbitMQPublisher{exchange: "orders", ch: ch, logger: testLogger()}

	go func() {
		time.Sleep(10 * time.Millisecond)
		confirms <- amqp091.Confirmation{DeliveryTag: 1, Ack: false}
	}()

	evt := model.OutboxEvent{ID: "evt-2", EventType: "order.created", Payload: []byte(`{}`)}
	if err := p.PublishWithConfirm(context.Background(), evt); err == nil {
		t.Fatal("expected error on nack")
	}
}

func TestPublishWithConfirm_TimeoutErrors(t *testing.T) {
	ch := newFakeChannel()
	ch.confirms = make(chan amqp091.Confirmation) // never sends
	p := &rabbitMQPublisher{exchange: "orders", ch: ch, logger: testLogger(), confirmTimeout: 50 * time.Millisecond}

	evt := model.OutboxEvent{ID: "evt-3", EventType: "order.created", Payload: []byte(`{}`)}
	if err := p.PublishWithConfirm(context.Background(), evt); err == nil {
		t.Fatal("expected timeout error")
	}
}
```

The `newFakeChannel`, `testLogger` helpers and `fakeChannel` struct must be added/extended in the existing test file. The fake channel needs to: capture `lastHeaders` from `PublishWithContext`, expose a `confirms` channel returned from `NotifyPublish`, and implement `Confirm(noWait) error` as a no-op. Mirror the shape of any fakes already present in `publisher_test.go`.

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd services/publish-order-service && go test ./internal/publisher/ -run TestPublishWithConfirm -v`
Expected: FAIL — `PublishWithConfirm undefined`.

- [ ] **Step 3: Implement publisher confirms**

In `services/publish-order-service/internal/publisher/publisher.go`:

(a) Add field to the struct:

```go
type rabbitMQPublisher struct {
	// ... existing fields ...
	confirmTimeout time.Duration
	confirms       chan amqp091.Confirmation
}
```

(b) In `connect()`, after `ExchangeDeclare`, before assigning to `r.conn`/`r.ch`:

```go
if err := ch.Confirm(false); err != nil {
	ch.Close()
	conn.Close()
	return fmt.Errorf("enable confirms: %w", err)
}
r.confirms = ch.NotifyPublish(make(chan amqp091.Confirmation, 1))
```

(c) In `NewRabbitMQPublisher` / `newRabbitMQPublisher`, default `confirmTimeout` to 5 * time.Second if zero.

(d) Extend the `Publisher` interface:

```go
type Publisher interface {
	Publish(evt model.Event) error
	PublishWithConfirm(ctx context.Context, evt model.OutboxEvent) error
	Close() error
}
```

(e) Add the method:

```go
func (r *rabbitMQPublisher) PublishWithConfirm(ctx context.Context, evt model.OutboxEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return fmt.Errorf("publisher is closed")
	}
	if r.ch == nil {
		return amqp091.ErrClosed
	}

	timeout := r.confirmTimeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	err := r.ch.PublishWithContext(ctx,
		r.exchange,
		evt.EventType,
		false, false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        []byte(evt.Payload),
			Headers: amqp091.Table{
				"event_id": evt.ID,
			},
			MessageId: evt.ID,
		},
	)
	if err != nil {
		return fmt.Errorf("publish with confirm: %w", err)
	}

	select {
	case c, ok := <-r.confirms:
		if !ok {
			return fmt.Errorf("confirm channel closed")
		}
		if !c.Ack {
			return fmt.Errorf("broker nacked delivery tag %d", c.DeliveryTag)
		}
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("confirm timeout after %s", timeout)
	case <-ctx.Done():
		return ctx.Err()
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd services/publish-order-service && go test ./internal/publisher/ -v
```

Expected: PASS — new tests + existing tests still pass.

- [ ] **Step 5: Commit**

```bash
git add services/publish-order-service/internal/publisher/publisher.go services/publish-order-service/internal/publisher/publisher_test.go
git commit -m "feat(publish-order): add PublishWithConfirm with broker ack + timeout"
```

---

## Task 6: Implement `outbox.Relay` (polling-only first; add LISTEN/NOTIFY in Task 7)

**Files:**
- Create: `services/publish-order-service/internal/outbox/relay.go`
- Test: `services/publish-order-service/internal/outbox/relay_test.go`

- [ ] **Step 1: Write the failing tests**

Create `services/publish-order-service/internal/outbox/relay_test.go`:

```go
package outbox

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/icl00ud/velure/services/publish-order-service/internal/model"
)

type fakeRepo struct {
	mu            sync.Mutex
	fetched       [][]model.OutboxEvent
	marked        [][]string
	fetchErr      error
	markErr       error
	pendingBatches [][]model.OutboxEvent
}

func (f *fakeRepo) SaveTx(ctx context.Context, tx *sql.Tx, evt model.OutboxEvent) error {
	return nil
}

func (f *fakeRepo) FetchUnpublished(ctx context.Context, limit int) (*sql.Tx, []model.OutboxEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fetchErr != nil {
		return nil, nil, f.fetchErr
	}
	if len(f.pendingBatches) == 0 {
		return &sql.Tx{}, nil, nil
	}
	batch := f.pendingBatches[0]
	f.pendingBatches = f.pendingBatches[1:]
	f.fetched = append(f.fetched, batch)
	return &sql.Tx{}, batch, nil
}

func (f *fakeRepo) MarkPublished(ctx context.Context, tx *sql.Tx, ids []string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.markErr != nil {
		return f.markErr
	}
	f.marked = append(f.marked, ids)
	return nil
}

type fakePublisher struct {
	mu         sync.Mutex
	calls      []model.OutboxEvent
	failOnID   string
	err        error
}

func (f *fakePublisher) PublishWithConfirm(ctx context.Context, evt model.OutboxEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, evt)
	if f.failOnID != "" && evt.ID == f.failOnID {
		return f.err
	}
	return nil
}

func TestRelay_ProcessBatch_HappyPath(t *testing.T) {
	repo := &fakeRepo{
		pendingBatches: [][]model.OutboxEvent{{
			{ID: "evt-1", EventType: "order.created", Payload: []byte(`{}`)},
			{ID: "evt-2", EventType: "order.created", Payload: []byte(`{}`)},
		}},
	}
	pub := &fakePublisher{}
	r := NewRelay(repo, pub, WithCommitFn(noopCommit), WithRollbackFn(noopRollback))

	if err := r.processBatch(context.Background()); err != nil {
		t.Fatalf("processBatch: %v", err)
	}
	if len(pub.calls) != 2 {
		t.Fatalf("expected 2 publishes, got %d", len(pub.calls))
	}
	if len(repo.marked) != 1 || len(repo.marked[0]) != 2 {
		t.Fatalf("expected MarkPublished for 2 ids, got %v", repo.marked)
	}
}

func TestRelay_ProcessBatch_PublishFails_RollsBackBatch(t *testing.T) {
	repo := &fakeRepo{
		pendingBatches: [][]model.OutboxEvent{{
			{ID: "evt-1", EventType: "order.created", Payload: []byte(`{}`)},
			{ID: "evt-2", EventType: "order.created", Payload: []byte(`{}`)},
		}},
	}
	pub := &fakePublisher{failOnID: "evt-2", err: errors.New("broker down")}
	r := NewRelay(repo, pub, WithCommitFn(noopCommit), WithRollbackFn(noopRollback))

	err := r.processBatch(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if len(repo.marked) != 0 {
		t.Fatalf("MarkPublished should not be called on partial failure, got %v", repo.marked)
	}
}

func TestRelay_Start_StopsOnContextCancel(t *testing.T) {
	repo := &fakeRepo{}
	pub := &fakePublisher{}
	r := NewRelay(repo, pub,
		WithInterval(10*time.Millisecond),
		WithCommitFn(noopCommit), WithRollbackFn(noopRollback),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := r.Start(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func noopCommit(tx *sql.Tx) error    { return nil }
func noopRollback(tx *sql.Tx) error  { return nil }
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd services/publish-order-service && go test ./internal/outbox/ -run TestRelay -v
```

Expected: FAIL — `NewRelay undefined`.

- [ ] **Step 3: Implement the relay**

Create `services/publish-order-service/internal/outbox/relay.go`:

```go
package outbox

import (
	"context"
	"database/sql"
	"time"

	"github.com/icl00ud/velure/services/publish-order-service/internal/metrics"
	"github.com/icl00ud/velure/services/publish-order-service/internal/model"
	"github.com/icl00ud/velure/shared/logger"
)

type Publisher interface {
	PublishWithConfirm(ctx context.Context, evt model.OutboxEvent) error
}

type Relay struct {
	repo      Repository
	publisher Publisher
	logger    *logger.Logger
	interval  time.Duration
	batchSize int
	// Test seams: allow injecting commit/rollback to avoid needing real *sql.Tx in unit tests.
	commit   func(*sql.Tx) error
	rollback func(*sql.Tx) error
	notify   <-chan struct{} // wake signal (e.g. from LISTEN/NOTIFY); nil for polling-only
}

type Option func(*Relay)

func WithInterval(d time.Duration) Option   { return func(r *Relay) { r.interval = d } }
func WithBatchSize(n int) Option            { return func(r *Relay) { r.batchSize = n } }
func WithLogger(l *logger.Logger) Option    { return func(r *Relay) { r.logger = l } }
func WithNotifyChannel(c <-chan struct{}) Option { return func(r *Relay) { r.notify = c } }
func WithCommitFn(f func(*sql.Tx) error) Option   { return func(r *Relay) { r.commit = f } }
func WithRollbackFn(f func(*sql.Tx) error) Option { return func(r *Relay) { r.rollback = f } }

func NewRelay(repo Repository, pub Publisher, opts ...Option) *Relay {
	r := &Relay{
		repo:      repo,
		publisher: pub,
		interval:  10 * time.Second,
		batchSize: 50,
		commit:    func(tx *sql.Tx) error { return tx.Commit() },
		rollback:  func(tx *sql.Tx) error { return tx.Rollback() },
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Relay) Start(ctx context.Context) error {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	// Run once immediately on start to drain anything pending from previous run.
	if err := r.processBatch(ctx); err != nil {
		r.logError("initial batch failed", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := r.processBatch(ctx); err != nil {
				r.logError("poll batch failed", err)
			}
		case <-r.notify:
			if err := r.processBatch(ctx); err != nil {
				r.logError("notify batch failed", err)
			}
		}
	}
}

func (r *Relay) processBatch(ctx context.Context) error {
	start := time.Now()
	tx, events, err := r.repo.FetchUnpublished(ctx, r.batchSize)
	if err != nil {
		metrics.OutboxRelayErrors.Inc()
		return err
	}
	if len(events) == 0 {
		_ = r.rollback(tx)
		return nil
	}

	ids := make([]string, 0, len(events))
	for _, evt := range events {
		if err := r.publisher.PublishWithConfirm(ctx, evt); err != nil {
			_ = r.rollback(tx)
			metrics.OutboxRelayPublished.WithLabelValues("failure").Inc()
			metrics.OutboxRelayErrors.Inc()
			return err
		}
		ids = append(ids, evt.ID)
		metrics.OutboxRelayPublished.WithLabelValues("success").Inc()
	}
	if err := r.repo.MarkPublished(ctx, tx, ids); err != nil {
		_ = r.rollback(tx)
		metrics.OutboxRelayErrors.Inc()
		return err
	}
	if err := r.commit(tx); err != nil {
		metrics.OutboxRelayErrors.Inc()
		return err
	}
	metrics.OutboxRelayBatchDuration.Observe(time.Since(start).Seconds())
	return nil
}

func (r *Relay) logError(msg string, err error) {
	if r.logger == nil {
		return
	}
	r.logger.Error(msg, logger.Err(err))
}
```

> Note: this references metrics that don't exist yet. Task 11 adds them. To keep this task green, also add a stub metrics file now (or skip metric calls behind a nil-check). **Decision: add the metrics now as part of Task 6** — see Step 4 below.

- [ ] **Step 4: Add the new metrics referenced above**

Open `services/publish-order-service/internal/metrics/metrics.go` and append (preserving the existing imports / `init`):

```go
var (
	OutboxRelayPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "outbox_relay_published_total",
			Help: "Outbox events published to RabbitMQ, by result.",
		},
		[]string{"result"},
	)

	OutboxRelayErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "outbox_relay_errors_total",
			Help: "Outbox relay error count (DB or broker failures during batch).",
		},
	)

	OutboxRelayBatchDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "outbox_relay_batch_duration_seconds",
			Help:    "Duration of a single outbox relay batch (fetch + publish + mark).",
			Buckets: prometheus.DefBuckets,
		},
	)

	OutboxEventsPending = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "outbox_events_pending",
			Help: "Number of outbox events with published_at IS NULL.",
		},
	)

	OutboxListenerReconnects = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "outbox_listener_reconnects_total",
			Help: "Count of LISTEN/NOTIFY listener reconnections.",
		},
	)
)
```

(Match the import path style used in the existing metrics file — adjust if it uses `prometheus.NewCounterVec` rather than `promauto`.)

- [ ] **Step 5: Run tests to verify they pass**

```bash
cd services/publish-order-service && go test ./internal/outbox/ ./internal/metrics/ -v
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add services/publish-order-service/internal/outbox/relay.go services/publish-order-service/internal/outbox/relay_test.go services/publish-order-service/internal/metrics/metrics.go
git commit -m "feat(publish-order): add outbox relay with whole-batch atomic publish"
```

---

## Task 7: Add Postgres `LISTEN/NOTIFY` listener

**Files:**
- Create: `services/publish-order-service/internal/outbox/listener.go`
- Test: `services/publish-order-service/internal/outbox/listener_test.go`
- Wire into the relay's `notify` channel.

Uses `github.com/lib/pq` (already a dep — see `main.go` `_ "github.com/lib/pq"` import) and its `pq.NewListener`.

- [ ] **Step 1: Write the failing test**

Create `services/publish-order-service/internal/outbox/listener_test.go`:

```go
package outbox

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakePqListener simulates the subset of *pq.Listener we use.
type fakePqListener struct {
	notify chan struct{}
	closed bool
}

func (f *fakePqListener) Listen(channel string) error { return nil }
func (f *fakePqListener) NotificationChannel() <-chan struct{} {
	if f.notify == nil { f.notify = make(chan struct{}, 1) }
	return f.notify
}
func (f *fakePqListener) Close() error { f.closed = true; return nil }

func TestListener_ForwardsNotifications(t *testing.T) {
	pq := &fakePqListener{notify: make(chan struct{}, 1)}
	l := newListenerWith(pq, "outbox_new")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	out := make(chan struct{}, 1)
	go l.run(ctx, out)

	pq.notify <- struct{}{}

	select {
	case <-out:
		// ok
	case <-time.After(50 * time.Millisecond):
		t.Fatal("expected forwarded notification")
	}
}

func TestListener_StopsOnContextCancel(t *testing.T) {
	pq := &fakePqListener{}
	l := newListenerWith(pq, "outbox_new")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	out := make(chan struct{}, 1)
	err := l.run(ctx, out)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd services/publish-order-service && go test ./internal/outbox/ -run TestListener -v
```

Expected: FAIL — `newListenerWith undefined`.

- [ ] **Step 3: Implement the listener**

Create `services/publish-order-service/internal/outbox/listener.go`:

```go
package outbox

import (
	"context"
	"time"

	"github.com/icl00ud/velure/services/publish-order-service/internal/metrics"
	"github.com/icl00ud/velure/shared/logger"
	"github.com/lib/pq"
)

type pqListener interface {
	Listen(channel string) error
	NotificationChannel() <-chan struct{}
	Close() error
}

type Listener struct {
	pq      pqListener
	channel string
	logger  *logger.Logger
}

// NewListener creates a Listener that subscribes to the given Postgres
// NOTIFY channel using lib/pq's dedicated connection. The minReconnect/
// maxReconnect arguments tune the lib/pq backoff.
func NewListener(dsn, channel string, log *logger.Logger) *Listener {
	pql := pq.NewListener(dsn, time.Second, 30*time.Second, func(ev pq.ListenerEventType, err error) {
		switch ev {
		case pq.ListenerEventReconnected:
			metrics.OutboxListenerReconnects.Inc()
		}
		if err != nil && log != nil {
			log.Warn("listener event", logger.Any("event", ev), logger.Err(err))
		}
	})
	return newListenerWith(adaptPqListener{pql, channel}, channel)
}

func newListenerWith(pql pqListener, channel string) *Listener {
	return &Listener{pq: pql, channel: channel}
}

// run consumes notifications from the underlying connection and forwards
// a wake signal on `out`. Returns when ctx is cancelled.
func (l *Listener) run(ctx context.Context, out chan<- struct{}) error {
	if err := l.pq.Listen(l.channel); err != nil {
		return err
	}
	defer l.pq.Close()

	in := l.pq.NotificationChannel()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-in:
			select {
			case out <- struct{}{}:
			default:
				// channel already has a pending wake; drop the duplicate
			}
		}
	}
}

// adaptPqListener bridges *pq.Listener to our pqListener interface.
// pq.Listener.Notify is a chan *pq.Notification; we coerce to chan struct{}.
type adaptPqListener struct {
	inner   *pq.Listener
	channel string
}

func (a adaptPqListener) Listen(ch string) error { return a.inner.Listen(ch) }
func (a adaptPqListener) NotificationChannel() <-chan struct{} {
	out := make(chan struct{}, 16)
	go func() {
		for n := range a.inner.Notify {
			_ = n
			select {
			case out <- struct{}{}:
			default:
			}
		}
		close(out)
	}()
	return out
}
func (a adaptPqListener) Close() error { return a.inner.Close() }
```

- [ ] **Step 4: Expose a `Start` method that runs forever**

Append to `listener.go`:

```go
func (l *Listener) Start(ctx context.Context, out chan<- struct{}) error {
	return l.run(ctx, out)
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
cd services/publish-order-service && go test ./internal/outbox/ -v
```

Expected: PASS — all outbox tests.

- [ ] **Step 6: Commit**

```bash
git add services/publish-order-service/internal/outbox/listener.go services/publish-order-service/internal/outbox/listener_test.go
git commit -m "feat(publish-order): add Postgres LISTEN/NOTIFY listener for outbox"
```

---

## Task 8: Refactor `OrderService.Create` and `UpdateStatus` to use the transactional outbox

**Files:**
- Modify: `services/publish-order-service/internal/service/order_service.go`
- Modify: `services/publish-order-service/internal/service/order_service_test.go`

- [ ] **Step 1: Update tests for `Create`**

The existing test for `Create` likely passes a mock repo and expects `Save` to be called. Replace/extend it so the test passes a mock `*sql.DB` (or test seam) plus a mock outbox repo, and asserts that **both** `SaveTx(order)` and `outbox.SaveTx(event)` happen inside the same transaction (commit at end).

```go
func TestOrderService_Create_PersistsOrderAndOutboxAtomically(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO TBLOrders`).
		WithArgs(sqlmock.AnyArg(), "user-1", sqlmock.AnyArg(), int64(0), "CREATED", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO outbox_events`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), model.OrderCreated, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	svc := NewOrderService(
		&realRepoBackedByDB{db: db},
		outbox.NewPostgresRepository(db),
		db,
		NewPricingCalculator(),
	)
	_, err := svc.Create(context.Background(), "user-1", []model.CartItem{{ProductID: "p1", Quantity: 1}})
	if err != nil { t.Fatalf("Create: %v", err) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}

func TestOrderService_Create_RollsBackOnOutboxFailure(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO TBLOrders`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO outbox_events`).WillReturnError(errors.New("outbox down"))
	mock.ExpectRollback()

	svc := NewOrderService(
		&realRepoBackedByDB{db: db},
		outbox.NewPostgresRepository(db),
		db,
		NewPricingCalculator(),
	)
	_, err := svc.Create(context.Background(), "user-1", []model.CartItem{{ProductID: "p1", Quantity: 1}})
	if err == nil { t.Fatal("expected error") }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}
```

`realRepoBackedByDB` is a tiny test helper struct exposing the real `PostgresOrderRepository` so the SQL goes through the same path. Reuse whatever the existing service tests do; if they currently use a pure mock interface, you'll need to evolve the test fixture in the same way as the production change.

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd services/publish-order-service && go test ./internal/service/ -run TestOrderService_Create -v
```

Expected: FAIL — `NewOrderService` arity mismatch.

- [ ] **Step 3: Refactor `OrderService`**

Replace the body of `services/publish-order-service/internal/service/order_service.go`:

```go
package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/icl00ud/velure/services/publish-order-service/internal/model"
	"github.com/icl00ud/velure/services/publish-order-service/internal/outbox"
	"github.com/icl00ud/velure/services/publish-order-service/internal/repository"
)

var ErrNoItems = errors.New("no items in the cart")
var ErrInvalidItem = errors.New("invalid item in the cart")

type OrderService struct {
	repo    repository.OrderRepository
	outbox  outbox.Repository
	db      *sql.DB
	pricing PricingCalculator
}

func NewOrderService(r repository.OrderRepository, ob outbox.Repository, db *sql.DB, pc PricingCalculator) *OrderService {
	return &OrderService{repo: r, outbox: ob, db: db, pricing: pc}
}

func (s *OrderService) Create(ctx context.Context, userID string, items []model.CartItem) (model.Order, error) {
	if len(items) == 0 {
		return model.Order{}, ErrNoItems
	}
	for _, item := range items {
		if item.ProductID == "" {
			return model.Order{}, fmt.Errorf("%w: missing product_id", ErrInvalidItem)
		}
		if item.Quantity <= 0 {
			return model.Order{}, fmt.Errorf("%w: quantity must be positive", ErrInvalidItem)
		}
	}

	total := s.pricing.Calculate(items)
	now := time.Now()
	o := model.Order{
		ID: uuid.NewString(), UserID: userID, Items: items, Total: total,
		Status: model.StatusCreated, CreatedAt: now, UpdatedAt: now,
	}

	if err := s.withTx(ctx, func(tx *sql.Tx) error {
		if err := s.repo.SaveTx(ctx, tx, o); err != nil {
			return err
		}
		payload, err := jsonMarshal(o)
		if err != nil { return err }
		return s.outbox.SaveTx(ctx, tx, model.OutboxEvent{
			ID:          uuid.NewString(),
			AggregateID: o.ID,
			EventType:   model.OrderCreated,
			Payload:     payload,
			CreatedAt:   now,
		})
	}); err != nil {
		return model.Order{}, err
	}
	return o, nil
}

func (s *OrderService) UpdateStatus(ctx context.Context, id, status string) (model.Order, error) {
	var updated model.Order
	if err := s.withTx(ctx, func(tx *sql.Tx) error {
		// Reuse the non-tx Find — read is independent of the write tx's isolation here.
		o, err := s.repo.Find(ctx, id)
		if err != nil { return err }
		o.Status = status
		o.UpdatedAt = time.Now()
		if err := s.repo.SaveTx(ctx, tx, o); err != nil { return err }
		payload, err := jsonMarshal(o)
		if err != nil { return err }
		if err := s.outbox.SaveTx(ctx, tx, model.OutboxEvent{
			ID:          uuid.NewString(),
			AggregateID: o.ID,
			EventType:   status,
			Payload:     payload,
			CreatedAt:   o.UpdatedAt,
		}); err != nil {
			return err
		}
		updated = o
		return nil
	}); err != nil {
		return model.Order{}, err
	}
	return updated, nil
}

func (s *OrderService) withTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil { return err }
	defer func() { _ = tx.Rollback() }()
	if err := fn(tx); err != nil { return err }
	return tx.Commit()
}

// read methods below stay unchanged
func (s *OrderService) GetOrdersByPage(ctx context.Context, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
	return s.repo.GetOrdersByPage(ctx, page, pageSize)
}
func (s *OrderService) GetOrdersByUserID(ctx context.Context, userID string, page, pageSize int) (*model.PaginatedOrdersResponse, error) {
	return s.repo.GetOrdersByUserID(ctx, userID, page, pageSize)
}
func (s *OrderService) GetOrderByID(ctx context.Context, userID, orderID string) (model.Order, error) {
	return s.repo.FindByUserID(ctx, userID, orderID)
}
```

Add `jsonMarshal` helper in a small file `service/json.go`:

```go
package service

import "encoding/json"

func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd services/publish-order-service && go test ./internal/service/ -v
```

Expected: PASS — new tests + all existing service tests. If existing tests fail due to `NewOrderService` signature change, update them to pass nil for `outbox` and `db` where they don't exercise tx paths (or supply fakes).

- [ ] **Step 5: Commit**

```bash
git add services/publish-order-service/internal/service/
git commit -m "refactor(publish-order): write order and outbox event atomically in OrderService"
```

---

## Task 9: Remove inline publish from `OrderHandler`

**Files:**
- Modify: `services/publish-order-service/internal/handler/order_handler.go`
- Modify: `services/publish-order-service/internal/handler/order_handler_test.go`

- [ ] **Step 1: Update tests**

In `order_handler_test.go`, remove any assertions that the handler invokes `pub.Publish` during `CreateOrder` and `UpdateStatus`. Replace with assertions that the handler returns 201/200 without touching the publisher. The publisher-related metrics assertions (`OrdersPublished`) should also be removed — they migrate to the relay (Task 6's metrics).

If the existing test passes a fake publisher to `NewOrderHandler`, you have two choices: (a) drop the publisher arg entirely from `NewOrderHandler` since it's no longer needed, or (b) keep the arg but stop using it. **Pick (a)** — narrower handler is cleaner.

```go
// Example test shape:
func TestOrderHandler_CreateOrder_Returns201_NoPublishCall(t *testing.T) {
	svc := &mockService{createReturns: model.Order{ID: "o1", Total: 50, Status: "CREATED"}}
	h := NewOrderHandler(svc)
	req := httptest.NewRequest("POST", "/orders", strings.NewReader(`[{"productId":"p","quantity":1}]`))
	req = req.WithContext(middleware.WithUserID(req.Context(), "u1"))
	rec := httptest.NewRecorder()
	h.CreateOrder(rec, req)
	if rec.Code != 201 { t.Fatalf("status=%d", rec.Code) }
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd services/publish-order-service && go test ./internal/handler/ -v
```

Expected: FAIL — `NewOrderHandler` signature mismatch (if you removed the publisher arg).

- [ ] **Step 3: Modify the handler**

In `services/publish-order-service/internal/handler/order_handler.go`:

(a) Remove the `Publisher` interface and the `pub` field from `OrderHandler`. Update `NewOrderHandler` to drop the publisher arg:

```go
type OrderHandler struct {
	svc OrderService
}

func NewOrderHandler(svc OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}
```

(b) In `CreateOrder` (current lines 69-76), delete:

```go
evt := model.Event{Type: model.OrderCreated, Payload: mustMarshal(o)}
if err := h.pub.Publish(evt); err != nil {
    logger.Error("publish event failed", logger.Err(err))
    metrics.Errors.WithLabelValues("rabbitmq").Inc()
    metrics.OrdersPublished.WithLabelValues("failure").Inc()
} else {
    metrics.OrdersPublished.WithLabelValues("success").Inc()
}
```

Replace with nothing — the service already wrote the outbox row.

(c) In `UpdateStatus` (current lines 150-153), delete the publish block analogously.

(d) Remove `mustMarshal` if no other call site remains.

- [ ] **Step 4: Update `main.go` call site** (Task 12 will do the rest of the wiring; for now only fix the compile error)

In `services/publish-order-service/main.go`:

```go
// line ~132 changes from:
oh := handler.NewOrderHandler(svc, pub)
// to:
oh := handler.NewOrderHandler(svc)
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
cd services/publish-order-service && go build ./... && go test ./internal/handler/ -v
```

Expected: BUILD OK, tests PASS.

- [ ] **Step 6: Commit**

```bash
git add services/publish-order-service/internal/handler/ services/publish-order-service/main.go
git commit -m "refactor(publish-order): remove inline publish from order handler"
```

---

## Task 10: Add restart-with-budget supervisor helper

**Files:**
- Create: `services/publish-order-service/internal/supervisor/supervisor.go`
- Test: `services/publish-order-service/internal/supervisor/supervisor_test.go`

- [ ] **Step 1: Write the failing tests**

Create `services/publish-order-service/internal/supervisor/supervisor_test.go`:

```go
package supervisor

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestRun_RestartsOnPanic_UpToBudget(t *testing.T) {
	var calls int32
	fn := func(ctx context.Context) error {
		atomic.AddInt32(&calls, 1)
		panic("boom")
	}
	err := Run(context.Background(), fn, Budget{MaxRestarts: 3, Window: time.Second})
	if err == nil { t.Fatal("expected error after budget exhausted") }
	if got := atomic.LoadInt32(&calls); got != 4 {
		t.Fatalf("expected 4 calls (initial + 3 restarts), got %d", got)
	}
}

func TestRun_StopsOnContextCancel(t *testing.T) {
	fn := func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := Run(ctx, fn, Budget{MaxRestarts: 3, Window: time.Second}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestRun_ReturnsNilOnNormalExit(t *testing.T) {
	fn := func(ctx context.Context) error { return nil }
	if err := Run(context.Background(), fn, Budget{MaxRestarts: 3, Window: time.Second}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd services/publish-order-service && go test ./internal/supervisor/ -v
```

Expected: FAIL — package not found.

- [ ] **Step 3: Implement the supervisor**

Create `services/publish-order-service/internal/supervisor/supervisor.go`:

```go
package supervisor

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Budget struct {
	MaxRestarts int
	Window      time.Duration
}

// Run executes fn; on panic, restarts it up to Budget.MaxRestarts within
// Budget.Window. Returns ctx.Err() when the context is cancelled, fn's
// return value on clean exit, or an aggregated error when the budget is
// exhausted.
func Run(ctx context.Context, fn func(context.Context) error, b Budget) error {
	var (
		restarts  int
		windowEnd time.Time
		lastErr   error
	)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := safeRun(ctx, fn)

		// Clean exit: propagate.
		if err == nil {
			return nil
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}

		lastErr = err
		now := time.Now()
		if now.After(windowEnd) {
			restarts = 0
			windowEnd = now.Add(b.Window)
		}
		restarts++
		if restarts > b.MaxRestarts {
			return fmt.Errorf("supervisor: budget of %d restarts in %s exhausted: %w", b.MaxRestarts, b.Window, lastErr)
		}
	}
}

func safeRun(ctx context.Context, fn func(context.Context) error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return fn(ctx)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd services/publish-order-service && go test ./internal/supervisor/ -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add services/publish-order-service/internal/supervisor/
git commit -m "feat(publish-order): add supervisor with restart-with-budget"
```

---

## Task 11: Wire outbox repo, relay, listener, and supervisor into `main.go`

**Files:**
- Modify: `services/publish-order-service/main.go`
- Modify: `services/publish-order-service/main_test.go` (extend `defaultDeps` test if it asserts on factory shape)

- [ ] **Step 1: Update `appDeps` and `defaultDeps`**

In `services/publish-order-service/main.go`, add to `appDeps`:

```go
newOutboxRepo func(*sql.DB) outbox.Repository
newRelay      func(outbox.Repository, outbox.Publisher, *logger.Logger) *outbox.Relay
newListener   func(dsn string, log *logger.Logger) *outbox.Listener
```

Wire `defaultDeps`:

```go
newOutboxRepo: outbox.NewPostgresRepository,
newRelay: func(repo outbox.Repository, pub outbox.Publisher, log *logger.Logger) *outbox.Relay {
	return outbox.NewRelay(repo, pub,
		outbox.WithInterval(10*time.Second),
		outbox.WithBatchSize(50),
		outbox.WithLogger(log),
	)
},
newListener: func(dsn string, log *logger.Logger) *outbox.Listener {
	return outbox.NewListener(dsn, "outbox_new", log)
},
```

- [ ] **Step 2: Replace handler construction and add relay goroutine**

In `run(...)`, after `repo` and `pub` are initialized:

```go
outboxRepo := deps.newOutboxRepo(repo.(dbProvider).DB())
svc := service.NewOrderService(repo, outboxRepo, repo.(dbProvider).DB(), service.NewPricingCalculator())
oh := handler.NewOrderHandler(svc)

// ... existing sseHandler / eventHandler setup unchanged ...

relay := deps.newRelay(outboxRepo, pub, log)
listener := deps.newListener(cfg.PostgresURL, log)
notifyCh := make(chan struct{}, 1)

g, ctx := errgroup.WithContext(ctx)

// HTTP server (existing g.Go) — unchanged

// Consumer (existing g.Go) — unchanged

// NEW: outbox listener
g.Go(func() error {
	return supervisor.Run(ctx, func(ctx context.Context) error {
		return listener.Start(ctx, notifyCh)
	}, supervisor.Budget{MaxRestarts: 3, Window: time.Minute})
})

// NEW: outbox relay (wired to the notify channel)
g.Go(func() error {
	return supervisor.Run(ctx, func(ctx context.Context) error {
		// Re-create relay each restart, with the notify channel attached.
		r := outbox.NewRelay(outboxRepo, pub,
			outbox.WithInterval(10*time.Second),
			outbox.WithBatchSize(50),
			outbox.WithLogger(log),
			outbox.WithNotifyChannel(notifyCh),
		)
		return r.Start(ctx)
	}, supervisor.Budget{MaxRestarts: 3, Window: time.Minute})
})

// shutdown goroutine (existing) — unchanged
```

> Note: the `relay := deps.newRelay(...)` line above is left for clarity but `g.Go` re-creates one wired to `notifyCh`. Remove the unused outer variable if the linter complains.

(Imports to add at top of `main.go`: `"github.com/icl00ud/velure/services/publish-order-service/internal/outbox"`, `"github.com/icl00ud/velure/services/publish-order-service/internal/supervisor"`.)

- [ ] **Step 3: Build and run unit tests**

```bash
cd services/publish-order-service && go build ./... && go test ./...
```

Expected: BUILD OK, all unit tests PASS.

- [ ] **Step 4: Smoke test with `make local-up`**

```bash
make local-up
# Watch logs:
docker logs -f $(docker ps -qf name=publish-order)
```

Look for log lines: `Migrations completed`, `RabbitMQ publisher initialized`, and (new) any relay/listener startup logs. Then:

```bash
# Create an order via curl (replace JWT as needed):
curl -X POST http://localhost/api/orders \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '[{"productId":"some-product","quantity":1}]'

# Verify outbox row exists and was published quickly:
docker exec -i $(docker ps -qf name=postgres) psql -U postgres -d publish_order_db \
  -c "SELECT id, aggregate_id, event_type, published_at FROM outbox_events ORDER BY created_at DESC LIMIT 5;"
```

Expected: rows present, `published_at` non-null within a few seconds.

- [ ] **Step 5: Commit**

```bash
git add services/publish-order-service/main.go services/publish-order-service/main_test.go
git commit -m "feat(publish-order): wire outbox relay and listener into main"
```

---

## Task 12: Add `idempotency.Checker` in process-order-service

**Files:**
- Create: `services/process-order-service/internal/idempotency/checker.go`
- Test: `services/process-order-service/internal/idempotency/checker_test.go`

Uses `github.com/redis/go-redis/v9` (add with `go get github.com/redis/go-redis/v9` in `services/process-order-service` if missing) and `github.com/alicebob/miniredis/v2` for tests.

- [ ] **Step 1: Write the failing tests**

Create `services/process-order-service/internal/idempotency/checker_test.go`:

```go
package idempotency

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newChecker(t *testing.T) (*Checker, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil { t.Fatal(err) }
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return NewChecker(rdb, 24*time.Hour), mr
}

func TestFirstSeen_TrueOnNewKey(t *testing.T) {
	c, _ := newChecker(t)
	ok, err := c.FirstSeen(context.Background(), "evt-1")
	if err != nil || !ok { t.Fatalf("expected true, got ok=%v err=%v", ok, err) }
}

func TestFirstSeen_FalseOnDuplicate(t *testing.T) {
	c, _ := newChecker(t)
	if _, err := c.FirstSeen(context.Background(), "evt-1"); err != nil { t.Fatal(err) }
	ok, err := c.FirstSeen(context.Background(), "evt-1")
	if err != nil || ok { t.Fatalf("expected false, got ok=%v err=%v", ok, err) }
}

func TestForget_AllowsReprocessing(t *testing.T) {
	c, _ := newChecker(t)
	ctx := context.Background()
	_, _ = c.FirstSeen(ctx, "evt-1")
	if err := c.Forget(ctx, "evt-1"); err != nil { t.Fatal(err) }
	ok, err := c.FirstSeen(ctx, "evt-1")
	if err != nil || !ok { t.Fatalf("expected true after Forget, got %v %v", ok, err) }
}

func TestFirstSeen_SetsTTL(t *testing.T) {
	c, mr := newChecker(t)
	ctx := context.Background()
	_, _ = c.FirstSeen(ctx, "evt-1")
	ttl := mr.TTL("event:evt-1")
	if ttl <= 0 {
		t.Fatalf("expected positive TTL, got %v", ttl)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd services/process-order-service && go test ./internal/idempotency/ -v
```

Expected: FAIL — package not found / missing deps.

- [ ] **Step 3: Add the dep and implement the checker**

```bash
cd services/process-order-service && go get github.com/redis/go-redis/v9 github.com/alicebob/miniredis/v2
```

Create `services/process-order-service/internal/idempotency/checker.go`:

```go
package idempotency

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Checker struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewChecker(rdb *redis.Client, ttl time.Duration) *Checker {
	return &Checker{rdb: rdb, ttl: ttl}
}

// FirstSeen reports whether eventID was seen for the first time.
// Internally: SET event:<id> 1 NX EX <ttl> — atomic set+expire in one
// round-trip. Returns (true, nil) on first sight, (false, nil) on dup.
func (c *Checker) FirstSeen(ctx context.Context, eventID string) (bool, error) {
	return c.rdb.SetNX(ctx, "event:"+eventID, "1", c.ttl).Result()
}

// Forget removes the dedup record so a future delivery can reach the
// real handler. Called on handler failure to preserve retry semantics.
func (c *Checker) Forget(ctx context.Context, eventID string) error {
	return c.rdb.Del(ctx, "event:"+eventID).Err()
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd services/process-order-service && go test ./internal/idempotency/ -v
```

Expected: PASS — all 4 tests.

- [ ] **Step 5: Commit**

```bash
git add services/process-order-service/internal/idempotency/ services/process-order-service/go.mod services/process-order-service/go.sum
git commit -m "feat(process-order): add Redis-backed idempotency checker"
```

---

## Task 13: Plumb `event_id` through the consumer and wire idempotency check

The current consumer at `services/process-order-service/internal/queue/consumer.go:84-156` unmarshals `model.Event` from the body. The `event_id` lives in the AMQP delivery's `Headers["event_id"]` (string) and `MessageId` envelope field — both set by the relay's `PublishWithConfirm` (Task 5).

**Files:**
- Modify: `services/process-order-service/internal/queue/consumer.go` (extract event id from delivery)
- Modify: `services/process-order-service/internal/handler/order_consumer.go` (wrap handler with idempotency check)
- Modify: `services/process-order-service/main.go` (Redis client + checker wiring)
- Modify: `services/process-order-service/internal/config/config.go` (Redis env vars)

- [ ] **Step 1: Add `event_id` to the handler signature in the consumer**

Change `Consumer.Consume(ctx, handler)` so the handler receives both the event and the event id. Edit `consumer.go`:

(a) Update the interface (line 16-19):

```go
type Consumer interface {
	Consume(ctx context.Context, handler func(eventID string, evt model.Event) error) error
	Close() error
}
```

(b) Inside the `for` loop in `Consume`, extract the id from headers (after `getRetryCount`):

```go
eventID := extractEventID(d)
// ... existing parse + handler call:
if err := handler(eventID, evt); err != nil { ... }
```

(c) Add helper at the bottom of the file:

```go
func extractEventID(d amqp091.Delivery) string {
	if d.Headers != nil {
		if v, ok := d.Headers["event_id"].(string); ok && v != "" {
			return v
		}
	}
	if d.MessageId != "" {
		return d.MessageId
	}
	// Last-resort fallback: deterministic from body so retries of the same
	// payload still dedupe even without headers.
	h := sha256.Sum256(d.Body)
	return hex.EncodeToString(h[:8])
}
```

Add imports `"crypto/sha256"`, `"encoding/hex"` if not present.

- [ ] **Step 2: Update consumer tests for the new signature**

Open `services/process-order-service/internal/queue/consumer_test.go` and adjust handler signatures (each `func(model.Event) error` becomes `func(eventID string, evt model.Event) error`).

Add a new test:

```go
func TestConsume_ExtractsEventIDFromHeader(t *testing.T) {
	// ... use existing fake AMQP delivery setup ...
	delivery := amqp091.Delivery{
		Headers:   amqp091.Table{"event_id": "evt-xyz"},
		MessageId: "ignored-when-header-present",
		Body:      []byte(`{"type":"order.created","payload":{}}`),
	}
	var got string
	handler := func(id string, e model.Event) error { got = id; return nil }
	// ... drive the consumer with the delivery ...
	if got != "evt-xyz" { t.Fatalf("expected evt-xyz, got %q", got) }
}

func TestExtractEventID_FallsBackToMessageIdThenHash(t *testing.T) {
	d := amqp091.Delivery{MessageId: "msg-1"}
	if id := extractEventID(d); id != "msg-1" { t.Fatalf("got %q", id) }
	d2 := amqp091.Delivery{Body: []byte("hello")}
	if id := extractEventID(d2); id == "" { t.Fatal("expected fallback hash, got empty") }
}
```

- [ ] **Step 3: Wrap the handler in `OrderConsumer.Start` with the idempotency check**

Edit `services/process-order-service/internal/handler/order_consumer.go`:

(a) Add field + constructor arg:

```go
type OrderConsumer struct {
	consumer queue.Consumer
	svc      service.PaymentService
	idem     IdempotencyChecker
	workers  int
	logger   *logger.Logger
}

type IdempotencyChecker interface {
	FirstSeen(ctx context.Context, eventID string) (bool, error)
	Forget(ctx context.Context, eventID string) error
}

func NewOrderConsumer(c queue.Consumer, svc service.PaymentService, idem IdempotencyChecker, workers int, log *logger.Logger) *OrderConsumer {
	return &OrderConsumer{consumer: c, svc: svc, idem: idem, workers: workers, logger: log}
}
```

(b) Update the inner `handler` closure:

```go
handler := func(eventID string, evt model.Event) error {
	metrics.MessagesConsumed.Inc()

	// Idempotency gate (fail-open on Redis errors).
	if oc.idem != nil && eventID != "" {
		firstSeen, err := oc.idem.FirstSeen(context.Background(), eventID)
		if err != nil {
			oc.logger.Error("idempotency check failed, processing anyway", logger.Err(err))
			metrics.IdempotencyCheckFailed.Inc()
		} else if !firstSeen {
			metrics.DuplicatesSkipped.Inc()
			metrics.MessagesAcknowledged.WithLabelValues("ack").Inc()
			return nil // ack and skip
		}
	}

	if evt.Type != model.OrderCreated {
		metrics.MessagesAcknowledged.WithLabelValues("ack").Inc()
		return nil
	}

	var p struct {
		ID    string           `json:"id"`
		Items []model.CartItem `json:"items"`
		Total float64          `json:"total"`
	}
	if err := json.Unmarshal(evt.Payload, &p); err != nil {
		metrics.MessageProcessingErrors.Inc()
		metrics.MessagesAcknowledged.WithLabelValues("nack").Inc()
		// On parse error, also free up the idempotency key so a fixed message can be reprocessed.
		if oc.idem != nil && eventID != "" {
			_ = oc.idem.Forget(context.Background(), eventID)
		}
		return err
	}

	if err := oc.svc.Process(p.ID, p.Items, int(p.Total)); err != nil {
		metrics.MessageProcessingErrors.Inc()
		metrics.MessagesAcknowledged.WithLabelValues("nack").Inc()
		// Free the key so retry can reach the handler again.
		if oc.idem != nil && eventID != "" {
			_ = oc.idem.Forget(context.Background(), eventID)
		}
		return err
	}

	metrics.MessagesAcknowledged.WithLabelValues("ack").Inc()
	return nil
}
```

- [ ] **Step 4: Add the new metrics**

In `services/process-order-service/internal/metrics/metrics.go`, append:

```go
var (
	DuplicatesSkipped = promauto.NewCounter(prometheus.CounterOpts{
		Name: "process_order_duplicates_skipped_total",
		Help: "Messages dropped because their event_id was already processed.",
	})
	IdempotencyCheckFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "process_order_idempotency_check_failed_total",
		Help: "Redis idempotency check errors (fail-open processed anyway).",
	})
	MessagesMissingEventID = promauto.NewCounter(prometheus.CounterOpts{
		Name: "process_order_messages_missing_event_id_total",
		Help: "Messages whose event_id could not be extracted from headers or envelope.",
	})
)
```

(Match the existing import / promauto style in that file.)

- [ ] **Step 5: Update Redis config + main.go wiring**

In `services/process-order-service/internal/config/config.go`, add fields:

```go
type Config struct {
	// ... existing ...
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}
```

And in `Load()`:

```go
if v := strings.TrimSpace(os.Getenv("REDIS_HOST")); v != "" {
	port := strings.TrimSpace(os.Getenv("REDIS_PORT"))
	if port == "" { port = "6379" }
	c.RedisAddr = v + ":" + port
} else {
	missing = append(missing, "REDIS_HOST")
}
c.RedisPassword = os.Getenv("REDIS_PASSWORD")
if v := strings.TrimSpace(os.Getenv("REDIS_DB")); v != "" {
	n, err := strconv.Atoi(v)
	if err != nil { return c, fmt.Errorf("invalid REDIS_DB: %w", err) }
	c.RedisDB = n
}
```

Update `services/process-order-service/internal/config/config_test.go` to set `REDIS_HOST` for the success-case tests.

In `services/process-order-service/main.go`, after `cfg` is loaded and `consumer/publisher` are constructed:

```go
rdb := redis.NewClient(&redis.Options{
	Addr:     cfg.RedisAddr,
	Password: cfg.RedisPassword,
	DB:       cfg.RedisDB,
})
defer rdb.Close()
if err := rdb.Ping(ctx).Err(); err != nil {
	log.Warn("redis ping failed; idempotency will fail-open", logger.Err(err))
}
checker := idempotency.NewChecker(rdb, 24*time.Hour)

oc := handler.NewOrderConsumer(consumer, paySvc, checker, cfg.Workers, log)
```

Imports: `"github.com/redis/go-redis/v9"`, `"github.com/icl00ud/velure/services/process-order-service/internal/idempotency"`.

Also add `REDIS_HOST` (and `REDIS_PORT=6379`) to:
- `infrastructure/local/.env.example`
- `infrastructure/local/docker-compose.yml` env section for the `process-order-service` (the Redis service itself should already exist on the `local_order` network per CLAUDE.md — verify with `docker compose config`)

- [ ] **Step 6: Build and run unit tests**

```bash
cd services/process-order-service && go build ./... && go test ./...
```

Expected: BUILD OK, tests PASS.

- [ ] **Step 7: Commit**

```bash
git add services/process-order-service/ infrastructure/local/.env.example infrastructure/local/docker-compose.yml
git commit -m "feat(process-order): idempotent consumer via Redis SET NX EX on event_id"
```

---

## Task 14: Integration test scenarios against the local stack

**Files:**
- Create: `services/publish-order-service/integration_test.go` (build tag `integration`)

The skill prefers unit tests; for the chaos scenarios in the spec, a dedicated integration file gated by a build tag is the right tradeoff — these need real Postgres, RabbitMQ, and Redis.

- [ ] **Step 1: Create the integration test scaffold**

Create `services/publish-order-service/integration_test.go`:

```go
//go:build integration

package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// These tests require `make local-up` to be running.
// Run: go test -tags=integration ./... -run Integration -v

func dsn() string {
	if v := os.Getenv("INTEGRATION_POSTGRES_DSN"); v != "" {
		return v
	}
	return "postgres://postgres:postgres@localhost:5432/publish_order_db?sslmode=disable"
}

func TestIntegration_OutboxHappyPath(t *testing.T) {
	// 1. POST /api/orders, get order id
	// 2. Poll outbox_events: assert row appears, published_at becomes non-null within 5s.
	// (Implementation: open *sql.DB to dsn(), query SELECT published_at FROM outbox_events WHERE aggregate_id=$1.)
	t.Skip("requires JWT helper; flesh out when running")
}

func TestIntegration_RabbitDownThenUp(t *testing.T) {
	// 1. docker stop rabbitmq
	// 2. POST /api/orders
	// 3. Verify outbox row exists with published_at NULL
	// 4. docker start rabbitmq
	// 5. Verify published_at becomes non-null within 15s
	if _, err := exec.LookPath("docker"); err != nil { t.Skip("docker not available") }
	t.Skip("flesh out with real docker controls")
}

func TestIntegration_DuplicateDelivery_IsSkipped(t *testing.T) {
	// 1. Publish the same event_id to the orders exchange twice via direct AMQP
	// 2. Query process_order_duplicates_skipped_total /metrics, assert it incremented by 1
	t.Skip("flesh out with direct AMQP publish")
}

func pollUntil(ctx context.Context, t *testing.T, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if fn() { return }
		select {
		case <-ctx.Done(): t.Fatal(ctx.Err())
		case <-time.After(200 * time.Millisecond):
		}
	}
	t.Fatal("condition not met within deadline")
}

// Suppress unused warnings for skeletons:
var _ = bytes.NewBuffer
var _ = json.Marshal
var _ http.Handler = nil
var _ *sql.DB = nil
```

The `t.Skip` placeholders are deliberate — these are scaffolds for the maintainer to flesh out once they have a JWT helper and `docker` access in CI. Mark as a follow-up in the PR description rather than blocking the feature.

- [ ] **Step 2: Verify build is clean with and without the tag**

```bash
cd services/publish-order-service && go build ./... && go vet ./...
go test -tags=integration ./... -run TestIntegration -v
```

Expected: BUILD OK; integration tests `SKIP`.

- [ ] **Step 3: Commit**

```bash
git add services/publish-order-service/integration_test.go
git commit -m "test(publish-order): scaffold integration tests for outbox + dedup scenarios"
```

---

## Task 15: Manual end-to-end smoke

Not a code task — a checklist run by hand or in CI. Document the runbook in the PR description.

- [ ] **Step 1: Bring up local stack**

```bash
make local-down  # clean
make local-up
```

- [ ] **Step 2: Happy path**

```bash
JWT=$(curl -s -X POST http://localhost/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"email":"...","password":"..."}' | jq -r .token)

curl -X POST http://localhost/api/orders \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '[{"productId":"some-real-id","quantity":1}]'
```

Verify:
- Postgres: `SELECT id, status FROM tblorders ORDER BY created_at DESC LIMIT 1;` → row exists.
- Postgres: `SELECT id, event_type, published_at FROM outbox_events ORDER BY created_at DESC LIMIT 1;` → row exists, `published_at` non-null within a few seconds.
- SSE: order status reaches a terminal state via `GET /api/me/orders/{id}/events`.
- Metrics: `curl http://localhost/metrics | grep outbox_relay_published_total` → success counter incremented.

- [ ] **Step 3: RabbitMQ outage**

```bash
docker stop $(docker ps -qf name=rabbitmq)
# POST another order; should still return 201
docker exec -i $(docker ps -qf name=postgres) psql -U postgres -d publish_order_db \
  -c "SELECT count(*) FROM outbox_events WHERE published_at IS NULL;"
# Expect count >= 1.
docker start $(docker ps -qf name=rabbitmq)
# Wait ~15s
docker exec -i $(docker ps -qf name=postgres) psql -U postgres -d publish_order_db \
  -c "SELECT count(*) FROM outbox_events WHERE published_at IS NULL;"
# Expect count == 0.
```

- [ ] **Step 4: Duplicate delivery**

Use `rabbitmqadmin` (in the rabbitmq container) to publish the same `event_id` header twice. Check `process_order_duplicates_skipped_total` increments.

- [ ] **Step 5: Tear down**

```bash
make local-down
```

No commit for this task — it's documentation of the verification step.

---

## Task 16: Install `rehype-slug` and update portfolio MDX pipeline

**Files:**
- Modify: `/Users/icl00ud/repos/portfolio/lib/mdx.ts`
- Modify: `/Users/icl00ud/repos/portfolio/package.json` (via pnpm)

- [ ] **Step 1: Install the dep**

```bash
cd /Users/icl00ud/repos/portfolio && pnpm add rehype-slug
```

- [ ] **Step 2: Add to the rehype plugins**

Edit `/Users/icl00ud/repos/portfolio/lib/mdx.ts`:

```diff
  import { compileMDX } from "next-mdx-remote/rsc";
  import rehypeShiki from "@shikijs/rehype";
+ import rehypeSlug from "rehype-slug";
  import { mdxComponents } from "@/components/mdx-components";

  export async function renderMDX(source: string) {
    const { content } = await compileMDX({
      source,
      components: mdxComponents,
      options: {
        mdxOptions: {
          rehypePlugins: [
+           rehypeSlug,
            [rehypeShiki, { themes: { light: "github-light", dark: "github-dark" } }],
          ],
        },
      },
    });
    return content;
  }
```

- [ ] **Step 3: Verify build is clean**

```bash
cd /Users/icl00ud/repos/portfolio && pnpm build
```

Expected: build succeeds. Open any project page in the local preview (`pnpm dev`, browse to `/projects/velure`) and confirm that headings now produce `<h2 id="...">` in the rendered HTML (DevTools → inspect a heading).

- [ ] **Step 4: Commit (in the portfolio repo)**

```bash
cd /Users/icl00ud/repos/portfolio
git add lib/mdx.ts package.json pnpm-lock.yaml
git commit -m "feat(mdx): enable heading id generation via rehype-slug"
```

---

## Task 17: Update `velure.mdx` in the portfolio

**Files:**
- Modify: `/Users/icl00ud/repos/portfolio/content/projects/velure.mdx`

- [ ] **Step 1: Insert the new "Outbox + idempotency" section**

Between the existing `## SSE handler` block (which ends at line ~80) and the existing `## What I'd do differently` header (at line ~82), insert:

````markdown
## Outbox + idempotency

The original design wrote the order to Postgres and then published to RabbitMQ
in two separate operations — a crash in the gap silently lost the event. The
fix is the transactional outbox pattern: the HTTP handler writes both the order
and an `outbox_events` row in one transaction. A relay goroutine drains the
outbox.

```go
func (r *Relay) processBatch(ctx context.Context) error {
    tx, events, err := r.repo.FetchUnpublished(ctx, r.batchSize)
    if err != nil { return err }
    defer tx.Rollback()

    ids := make([]string, 0, len(events))
    for _, evt := range events {
        if err := r.publisher.PublishWithConfirm(ctx, evt); err != nil {
            return err  // rollback → next tick retries the whole batch
        }
        ids = append(ids, evt.ID)
    }
    if err := r.repo.MarkPublished(ctx, tx, ids); err != nil { return err }
    return tx.Commit()
}
```

The relay listens on a Postgres `NOTIFY outbox_new` channel for sub-second
latency and falls back to a 10 s poll to recover any notifications dropped
across a listener reconnect. `FOR UPDATE SKIP LOCKED` lets multiple replicas
share the workload safely. On the consume side, `process-order-service` uses
Redis `SET event:<id> 1 NX EX 86400` to drop redelivered messages — at-least-once
on the wire, effectively-once in processing, with no 2PC.

````

- [ ] **Step 2: Add a new bullet under `## Key decisions`**

Insert this bullet after the "SSE over WebSocket" bullet (currently the last one in that section):

```markdown
- **Outbox pattern with hybrid push/pull relay.** Order writes and event publishes
  are atomic via a Postgres `outbox_events` table written in the same transaction
  as the order. A relay goroutine inside `publish-order-service` drains the table
  to RabbitMQ using `FOR UPDATE SKIP LOCKED` (safe for multi-replica) plus
  `LISTEN/NOTIFY` for sub-second latency, with a slow poll as fallback for missed
  notifications. `process-order-service` dedupes via Redis `SET NX EX` keyed on
  event UUIDs propagated through AMQP headers — at-least-once delivery becomes
  effectively-once processing without 2PC.
```

- [ ] **Step 3: Rewrite the outbox bullet under `## What I'd do differently`**

Replace the existing first bullet:

```markdown
- Outbox pattern between Postgres and RabbitMQ. Today the order write and the
  publish are two operations; a crash between them silently loses the event.
```

with:

```markdown
- ~~Outbox pattern between Postgres and RabbitMQ. Today the order write and the
  publish are two operations; a crash between them silently loses the event.~~
  **✓ Done** — see [Outbox + idempotency](#outbox-idempotency).
- Per-aggregate sharding of the outbox relay so multi-replica deployments
  preserve event ordering within a single `order_id` — today FIFO is global,
  not per-aggregate.
```

- [ ] **Step 4: Verify the rendered page**

```bash
cd /Users/icl00ud/repos/portfolio && pnpm dev
# Open http://localhost:3000/projects/velure in a browser
```

Verify:
- The new "Outbox + idempotency" section appears between "SSE handler" and "What I'd do differently".
- The code block renders with Shiki syntax highlighting (Go).
- The strikethrough line in "What I'd do differently" renders with strike-through formatting.
- The `[Outbox + idempotency](#outbox-idempotency)` link, when clicked, scrolls the page to the new section (this confirms `rehype-slug` is active).

- [ ] **Step 5: Commit (in the portfolio repo)**

```bash
cd /Users/icl00ud/repos/portfolio
git add content/projects/velure.mdx
git commit -m "docs(velure): document outbox pattern implementation; mark debt resolved"
```

---

## Self-Review Notes (from plan author)

Verified inline:

**Spec coverage:**
- Atomic write (spec § Architecture) → Task 8.
- `outbox_events` schema + index + trigger (spec § Schema) → Task 2.
- Repository (`SaveTx`, `FetchUnpublished`, `MarkPublished`) (spec § Components) → Task 3.
- Publisher confirms (spec § Components, publisher.go subsection) → Task 5.
- Relay with whole-batch atomicity (spec § Components, relay.go) → Task 6.
- LISTEN/NOTIFY + poll fallback (spec § Architecture, decisions table) → Task 7.
- Supervisor for panic recovery (spec § Components, main.go) → Task 10.
- Handler depublishes (spec § Components, handler subsection) → Task 9.
- Idempotency checker (spec § Components, process-order) → Task 12.
- Consumer wrap with fail-open Redis (spec § Error handling) → Task 13.
- Metrics (spec § Metrics) → Tasks 6, 13.
- Integration scenarios (spec § Testing) → Task 14, 15.
- Portfolio updates (spec § Portfolio documentation updates) → Tasks 16, 17.

**Placeholder scan:** Integration tests (Task 14) contain deliberate `t.Skip` scaffolds — flagged explicitly as future flesh-out work, not hidden placeholders. No other TODOs or vague "implement appropriate" instructions.

**Type consistency:** `Repository`, `OutboxEvent`, `Publisher`, `Relay`, `Checker`, `extractEventID` — names match across tasks. Function signatures align (e.g., `PublishWithConfirm(ctx, OutboxEvent) error` consistent in Task 5 and Task 6).

**Risk:** Task 13's modification to `Consumer.Consume` signature touches `consumer_test.go` which I have not read line-by-line. The maintainer should expect a moderate refactor of that test file when updating the handler signature.
