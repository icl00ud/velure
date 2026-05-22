# Outbox Pattern + Consumer Idempotency

**Date:** 2026-05-22
**Status:** Draft — pending implementation
**Scope:** `services/publish-order-service`, `services/process-order-service`, infra (Postgres migration, Redis client), portfolio doc update.

## Problem

The order write path performs two side effects:

```go
// services/publish-order-service/internal/handler/order_handler.go
o, err := h.svc.Create(r.Context(), userID, items)  // 1. INSERT into Postgres
// ...
evt := model.Event{Type: model.OrderCreated, Payload: mustMarshal(o)}
if err := h.pub.Publish(evt); err != nil {           // 2. AMQP publish
    logger.Error("publish event failed", logger.Err(err))
}
```

The same shape exists in `UpdateStatus` (lines 143 + 151). Postgres and RabbitMQ are independent resource managers; there is no atomic commit across them without 2PC (which is impractical). A crash between steps 1 and 2, or a RabbitMQ outage at step 2, leaves the order persisted with no event downstream — `process-order-service` never runs, the SSE stream never advances past `CREATED`, and the customer sees a stuck order. The handler currently returns `201 Created` even when the publish fails, so the client has no signal to retry.

The DLQ added in commit `ec58f57` protects the consume side from poison messages but does not address loss on the publish side.

## Goal

Guarantee that **every persisted order produces exactly one downstream effect**, even across process crashes, broker outages, and arbitrary network failures, without introducing 2PC.

Non-goals:
- HTTP-level request idempotency (duplicate POSTs from a retrying client creating duplicate orders). Tracked as future work; out of scope here.
- Backfill of historical orders (events flow only for new writes after deploy).

## Approach

Adopt the **transactional outbox pattern**: order writes and event records share one Postgres transaction. A relay goroutine drains the outbox to RabbitMQ asynchronously. The consumer dedupes redelivered messages via Redis. At-least-once on the wire becomes effectively-once in processing.

### Decisions (with alternatives rejected)

| Decision | Choice | Rejected alternatives |
|----------|--------|-----------------------|
| Scope | `Create` + `UpdateStatus` in `publish-order-service`; consumer idempotency in `process-order-service` | Only `Create` (leaves `UpdateStatus` inconsistent); also outbox in `process-order-service` (would add Postgres to a deliberately stateless service) |
| Relay placement | Goroutine inside `publish-order-service` | Separate `outbox-relay-service` (more YAML, no scaling benefit at this volume); Debezium/CDC (overkill, drops the pedagogical value of writing the pattern) |
| Consumer dedup | Redis `SET key value NX EX <ttl>` (atomic set+expire, one round-trip) | In-memory LRU (lost on pod restart, breaks across replicas); Postgres in `process-order-service` (would contradict the "no DB of its own" architectural decision) |
| Relay wake strategy | Postgres `LISTEN/NOTIFY` push + 10 s poll fallback | Pure 1 s polling (idle DB load, 500 ms avg latency); pure NOTIFY (NOTIFY is fire-and-forget — listener reconnect drops in-flight notifications) |
| Per-aggregate ordering across multi-replica relays | Documented as future work — single-replica relay sufficient for current scope | Consistent-hash sharding via `hashtext(aggregate_id) % N` (premature; project does not need it yet) |
| Redis outage policy | Fail-open (process the message, accept rare duplicates during outage) | Fail-closed (requeue/block) — risks stuck queue if Redis is down for hours |

## Architecture

```
[Antes]
HTTP POST /orders → publish-order
                      ├─ INSERT orders (Postgres)
                      └─ Publish RabbitMQ           ← inconsistency window

[Depois]
HTTP POST /orders → publish-order
                      └─ BEGIN TX
                           ├─ INSERT orders
                           └─ INSERT outbox_events
                         COMMIT                     ← atomic

[publish-order, relay goroutine]
  on NOTIFY outbox_new OR every 10 s poll OR ctx tick:
    BEGIN
      SELECT * FROM outbox_events
        WHERE published_at IS NULL
        ORDER BY created_at
        LIMIT 50
        FOR UPDATE SKIP LOCKED
      for each event:
        publish → RabbitMQ (publisher confirms)
      UPDATE outbox_events SET published_at = now() WHERE id = ANY($1)
    COMMIT

[process-order, consumer]
  on message:
    eventID := msg.Headers["event_id"]   (fallback: MessageId → sha256(body))
    firstSeen, _ := redis.SET("event:"+eventID, "1", "NX", "EX", 86400)
    if !firstSeen:
      Ack + metric DuplicatesSkipped++ + return
    process normally
    on handler error: redis.DEL("event:"+eventID), Nack(requeue=true)
    on handler ok: Ack
```

## Schema

New migration in `services/publish-order-service/internal/database/migrate.go`:

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

Rationale per column:

| Column | Choice |
|--------|--------|
| `id UUID` | Generated in app (`uuid.NewString()`) so the application knows the id before commit; reused as the AMQP `event_id` header and the Redis dedup key. |
| `aggregate_id` | Order id, for debug queries ("all events for order X"). No FK to `orders` — the outbox is an append log that must survive aggregate deletions. |
| `event_type` | String, not enum — schema changes do not require `ALTER TYPE`. Becomes the AMQP routing key. |
| `payload JSONB` | Snapshot of the order at event time. `JSONB` over `TEXT` for `jsonb_pretty` in debug queries. |
| `created_at` | FIFO ordering; partial index column. |
| `published_at` | Tri-state via NULL semantics — simpler than a status enum. |

The partial index ensures the relay's `WHERE published_at IS NULL` query stays cheap as the historical table grows. Publishing flips the row out of the index; backlog scans never touch published rows.

Published rows are retained for audit. A separate retention job (out of scope for this spec; document as ops runbook later) can `DELETE WHERE published_at < now() - interval '7 days'` if volume justifies it.

## Components

### New: `services/publish-order-service/internal/outbox/`

```go
// repository.go
type Repository interface {
    SaveTx(ctx context.Context, tx *sql.Tx, evt model.OutboxEvent) error
    FetchUnpublished(ctx context.Context, limit int) (*sql.Tx, []model.OutboxEvent, error)
    MarkPublished(ctx context.Context, tx *sql.Tx, ids []string) error
}
```

`FetchUnpublished` opens the transaction internally and returns it open because `FOR UPDATE SKIP LOCKED` only holds the lock while the transaction is alive. The relay is responsible for committing or rolling back. This intentionally couples lock lifetime to publish lifetime: any error path between fetch and `MarkPublished` results in a rollback, the lock releases, and the events are picked up on the next iteration. No partial state is ever persisted.

```go
// relay.go
type Relay struct {
    repo      Repository
    publisher Publisher           // PublishWithConfirm(evt) error
    listener  *pq.Listener        // NOTIFY subscriber
    interval  time.Duration       // poll fallback, default 10s
    batchSize int                 // default 50
}

func (r *Relay) Start(ctx context.Context) error {
    pollTicker := time.NewTicker(r.interval)
    defer pollTicker.Stop()
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-r.listener.Notify:
            r.processBatch(ctx)   // drain on push
        case <-pollTicker.C:
            r.processBatch(ctx)   // safety net for missed notifications
        }
    }
}

func (r *Relay) processBatch(ctx context.Context) error {
    tx, events, err := r.repo.FetchUnpublished(ctx, r.batchSize)
    if err != nil { return err }
    defer tx.Rollback()

    ids := make([]string, 0, len(events))
    for _, evt := range events {
        if err := r.publisher.PublishWithConfirm(evt); err != nil {
            return err  // rollback → next iteration retries the whole batch
        }
        ids = append(ids, evt.ID)
    }
    if err := r.repo.MarkPublished(ctx, tx, ids); err != nil { return err }
    return tx.Commit()
}
```

Key technical points:

- **Publisher confirms must be enabled** (`channel.Confirm(false)` + wait on the `confirms` channel with a 5 s timeout per message). Without this, a successful client-side `Publish` does not guarantee broker receipt. A confirm timeout is treated as a failure: rollback, retry next iteration, accept that the broker may have received the message anyway — the consumer dedup handles the duplicate.
- **Whole-batch atomicity**: a single publish failure rolls back the entire batch. This is intentional — partial marking is harder to reason about than retry-the-batch. Batch size 50 bounds the retry cost.
- **Listener connection management**: `pq.Listener` opens a dedicated Postgres connection (not from the pool — `LISTEN` requires a long-lived connection). It reconnects with backoff (1 s → 2 s → 4 s → 30 s max). On reconnect, the relay immediately runs one `processBatch` before returning to idle, to recover any `NOTIFY` events dropped during the disconnect.
- **Panic safety**: the loop body runs inside a `defer recover()` block. A panic logs at fatal and is allowed to crash the goroutine; `main.go` restarts it up to 3 times in 1 minute, after which it fails-fast and lets Kubernetes recreate the pod.

### Modified: `services/publish-order-service/internal/repository/order_repository.go`

`OrderRepository` interface gains a `SaveTx(ctx, tx, order)` method. The existing `Save(ctx, order)` remains for callers that do not need transactional semantics (`Find`, `Update` paths unchanged).

### Modified: `services/publish-order-service/internal/service/order_service.go`

`OrderService` now depends on `*sql.DB` (to open transactions) and `outbox.Repository` (in addition to the existing `repository.OrderRepository`). `Create` and `UpdateStatus` are rewritten as:

```go
func (s *OrderService) Create(ctx context.Context, userID string, items []model.CartItem) (model.Order, error) {
    // ... validation + order construction unchanged ...

    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil { return model.Order{}, err }
    defer tx.Rollback()  // no-op if Commit succeeded

    if err := s.orders.SaveTx(ctx, tx, o); err != nil {
        return model.Order{}, err
    }
    evt := model.OutboxEvent{
        ID:          uuid.NewString(),
        AggregateID: o.ID,
        EventType:   model.OrderCreated,
        Payload:     mustMarshal(o),
        CreatedAt:   now,
    }
    if err := s.outbox.SaveTx(ctx, tx, evt); err != nil {
        return model.Order{}, err
    }
    if err := tx.Commit(); err != nil {
        return model.Order{}, err
    }
    return o, nil
}
```

### Modified: `services/publish-order-service/internal/handler/order_handler.go`

The `h.pub.Publish(evt)` calls in `CreateOrder` (lines 69-76) and `UpdateStatus` (lines 150-153) are removed. The handler's responsibility shrinks to validation + delegation. `metrics.OrdersPublished` migrates from the handler to the relay (`outbox_relay_published_total`).

### New: `services/publish-order-service/internal/model/event.go` additions

```go
type OutboxEvent struct {
    ID           string
    AggregateID  string
    EventType    string
    Payload      json.RawMessage
    CreatedAt    time.Time
    PublishedAt  *time.Time
}
```

### Modified: `services/publish-order-service/internal/publisher/publisher.go`

Gains `PublishWithConfirm(evt model.OutboxEvent) error`. Internally enables publisher confirms on channel init, sets `event_id` AMQP header from `evt.ID`, waits on the confirm channel with a 5 s timeout. The existing `Publish(model.Event) error` method is removed once no callers remain.

### New: `services/publish-order-service/main.go` wiring

```go
relay := outbox.NewRelay(outboxRepo, publisher, listener,
    outbox.WithInterval(10*time.Second),
    outbox.WithBatchSize(50),
)
go func() {
    if err := supervisor.Run(ctx, relay.Start, supervisor.MaxRestarts(3, time.Minute)); err != nil {
        logger.Fatal("outbox relay failed permanently", logger.Err(err))
    }
}()
```

`supervisor` is a tiny helper (~30 LOC) added under `internal/supervisor/` to handle restart-with-budget — not a new dep.

### New: `services/process-order-service/internal/idempotency/`

```go
type Checker struct {
    rdb *redis.Client
    ttl time.Duration   // 24h default
}

func (c *Checker) FirstSeen(ctx context.Context, eventID string) (bool, error) {
    ok, err := c.rdb.SetNX(ctx, "event:"+eventID, "1", c.ttl).Result()
    return ok, err
}

func (c *Checker) Forget(ctx context.Context, eventID string) error {
    return c.rdb.Del(ctx, "event:"+eventID).Err()
}
```

`SetNX` in `go-redis/v9` issues `SET key value EX <ttl> NX` on the wire — set and expire are atomic in a single round-trip.

### Modified: `services/process-order-service/internal/consumer/rabbitmq_consumer.go`

The handler is wrapped:

```go
eventID := extractEventID(msg)  // header → MessageId → sha256(body)
firstSeen, err := idem.FirstSeen(ctx, eventID)
if err != nil {
    logger.Error("idempotency check failed, processing anyway", logger.Err(err))
    metrics.IdempotencyCheckFailed.Inc()
    firstSeen = true  // fail-open
}
if !firstSeen {
    msg.Ack(false)
    metrics.DuplicatesSkipped.Inc()
    return
}
if err := handler(ctx, msg); err != nil {
    _ = idem.Forget(ctx, eventID)   // allow retry to reach handler
    msg.Nack(false, true)            // requeue
    return
}
msg.Ack(false)
```

### Modified: `services/process-order-service/internal/config/config.go`

Adds `RedisHost`, `RedisPort`, `RedisPassword`, `RedisDB`. Redis is already present on the `local_order` Docker network (per `CLAUDE.md`); no compose change needed.

## Error handling matrix

| Failure | Behavior | Customer impact |
|---------|----------|-----------------|
| `BeginTx` fails | Handler returns 503 | Client retries |
| `SaveTx(order)` fails | Rollback, 500 | Client retries |
| `SaveTx(outbox)` fails | Rollback (includes order), 500 | Client retries — order is never persisted without its event |
| `Commit` fails | Auto-rollback, 500 | Client retries |
| Process dies after `Commit`, before HTTP 201 | Order + outbox persisted, client sees network error and may retry | Duplicate order possible — out of scope (HTTP idempotency is future work) |
| `FetchUnpublished` fails (DB blip) | Log warn, metric `outbox_relay_errors_total++`, next iteration retries | None — event still pending |
| RabbitMQ down at publish | Batch rollback, retry on next iteration | Latency only — event will publish when broker recovers |
| Publisher confirm times out | Treat as failure, rollback, retry | Possible broker-side duplicate, deduped by consumer |
| `MarkPublished` fails after confirm OK | Rollback — event will republish | Duplicate published, deduped by consumer |
| Process dies between confirm and `MarkPublished` | Same as above | Duplicate published, deduped by consumer |
| Relay panic | `recover()` + supervisor restarts up to 3× per minute, then fail-fast | Outage in event flow until restart succeeds or k8s recreates pod |
| Consumer: message has no `event_id` header (legacy) | Fallback to AMQP `MessageId`, then `sha256(payload)[:16]` | Best-effort dedup; metric `messages_missing_event_id_total++` |
| Consumer: Redis down | Fail-open — log error, process the message | Rare duplicates accepted during outage |
| Consumer: handler error after Redis mark | `DEL` the Redis key, `Nack(requeue=true)` | Retry reaches the real handler |

## Metrics

New Prometheus metrics:

- `outbox_events_pending` (gauge) — count of `published_at IS NULL`. Sampled by the relay every 30 s.
- `outbox_relay_batch_duration_seconds` (histogram).
- `outbox_relay_published_total{result="success|failure"}` (counter) — replaces `OrdersPublished` from the handler.
- `outbox_relay_errors_total` (counter).
- `outbox_listener_reconnects_total` (counter).
- `process_order_duplicates_skipped_total` (counter).
- `process_order_idempotency_check_failed_total` (counter).
- `process_order_messages_missing_event_id_total` (counter).

Suggested alert (document only, no Prometheus rule change in this scope): `outbox_events_pending > 1000` for 5 minutes.

## Testing

### Unit tests

- `outbox.Repository`: `SaveTx`, `FetchUnpublished`, `MarkPublished` against `sqlmock`. Verify `FOR UPDATE SKIP LOCKED` is emitted; verify partial index filter.
- `outbox.Relay`: mocked `Publisher` and `Repository`. Scenarios: empty batch, publish error mid-batch (assert rollback, no `MarkPublished` call), full success, panic recovery.
- `idempotency.Checker`: against `miniredis`. First-seen returns true; second call returns false; `Forget` removes key.

### Integration tests

Run against the existing `make local-up` stack (Postgres + RabbitMQ + Redis):

1. Happy path: `POST /orders` → poll `outbox_events` → verify row exists, then `published_at` set within 2 s.
2. RabbitMQ down mid-flight: `docker stop rabbitmq` between order creation and relay tick; verify `published_at IS NULL`. `docker start rabbitmq`; verify publish completes within 12 s (10 s poll fallback + slack).
3. Crash recovery: launch publish-order with a feature flag that calls `os.Exit(1)` after `Commit` but before HTTP response; restart; verify event publishes on next relay tick.
4. Duplicate delivery: publish the same `event_id` twice via direct AMQP; assert `process-order` handler invoked once, `DuplicatesSkipped` increments by 1.
5. Redis fail-open: stop Redis; publish event; assert handler still invoked, `IdempotencyCheckFailed` increments.

### E2E

Extend the existing end-to-end suite with a chaos scenario: kill RabbitMQ for 5 s in the middle of an order flow, assert the order reaches a terminal status within 30 s.

## Future work (explicitly out of scope)

- **HTTP idempotency** via `Idempotency-Key` header on `POST /orders` to handle client retry duplicates.
- **Per-aggregate sharding of the relay**: when multi-replica publish-order is needed, partition the outbox by `hashtext(aggregate_id) % N` so events for a single `order_id` always land on the same relay, preserving in-order delivery. Today, single-replica relay suffices and global FIFO is good enough.
- **Retention job** for published rows (`DELETE WHERE published_at < now() - interval '7 days'`).
- **Architecture diagram update** (`/cdn/architecture.gif`) to show the outbox table and relay loop.

## Portfolio documentation updates

Target: `/Users/icl00ud/repos/portfolio/content/projects/velure.mdx`.

### 1. Pipeline prerequisite

Add `rehype-slug` to the MDX pipeline so heading anchors resolve. Edit `lib/mdx.ts`:

```diff
  import { compileMDX } from "next-mdx-remote/rsc";
  import rehypeShiki from "@shikijs/rehype";
+ import rehypeSlug from "rehype-slug";

  rehypePlugins: [
+   rehypeSlug,
    [rehypeShiki, ...],
  ],
```

Plus `pnpm add rehype-slug`.

### 2. New section in `velure.mdx`, inserted between "SSE handler" and "What I'd do differently"

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
        if err := r.publisher.PublishWithConfirm(evt); err != nil {
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
on the wire, exactly-once in processing, with no 2PC.
````

### 3. New bullet in "Key decisions"

```markdown
- **Outbox pattern with hybrid push/pull relay.** Order writes and event publishes
  are atomic via a Postgres `outbox_events` table written in the same transaction
  as the order. A relay goroutine inside `publish-order-service` drains the table
  to RabbitMQ using `FOR UPDATE SKIP LOCKED` (safe for multi-replica) plus
  `LISTEN/NOTIFY` for sub-second latency, with a slow poll as fallback for
  missed notifications. `process-order-service` dedupes via Redis `SET NX EX`
  keyed on event UUIDs propagated through AMQP headers — at-least-once delivery
  becomes effectively-once processing without 2PC.
```

### 4. "What I'd do differently" — mark resolved + add new debt

```markdown
- ~~Outbox pattern between Postgres and RabbitMQ. Today the order write and the
  publish are two operations; a crash between them silently loses the event.~~
  **✓ Done** — see [Outbox + idempotency](#outbox-idempotency).
- Per-aggregate sharding of the outbox relay so multi-replica deployments
  preserve event ordering within a single `order_id` — today FIFO is global,
  not per-aggregate.
```

The strikethrough + checkmark preserves the original honesty about the gap and
shows the project's evolution. The link points the reader to the in-page section
explaining the implementation.

## Acceptance criteria

- All five integration test scenarios above pass.
- `metrics.Errors{type="rabbitmq"}` no longer increments on RabbitMQ outage during a `POST /orders` — instead, `outbox_events_pending` rises and drains automatically when the broker recovers.
- Restarting `publish-order-service` mid-flight does not lose events: any row with `published_at IS NULL` at restart is published within one relay cycle of startup.
- `process-order-service` invoked twice with the same `event_id` runs the real handler exactly once; `DuplicatesSkipped` increments.
- Portfolio site builds clean with `pnpm build`; heading anchor `#outbox-idempotency` resolves and the link from "What I'd do differently" scrolls to the new section.
