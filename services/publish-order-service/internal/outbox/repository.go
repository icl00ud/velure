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

	// CountPending returns the number of events not yet published.
	CountPending(ctx context.Context) (int64, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) SaveTx(ctx context.Context, tx *sql.Tx, evt model.OutboxEvent) error {
	const q = `
		INSERT INTO outbox_events (id, aggregate_id, event_type, payload, created_at, trace_context)
		VALUES ($1, $2, $3, $4, $5, $6);
	`
	created := evt.CreatedAt
	if created.IsZero() {
		created = time.Now().UTC()
	}
	_, err := tx.ExecContext(ctx, q, evt.ID, evt.AggregateID, evt.EventType, []byte(evt.Payload), created, evt.TraceContext)
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

	// Aggregate-level sharding: pg_try_advisory_xact_lock(hashtext(aggregate_id))
	// makes each relay replica claim whole aggregates for the duration of its
	// transaction. Another replica's fetch skips every event of a claimed
	// aggregate (the try-lock fails), so events of one order are never split
	// across replicas and per-aggregate ordering survives horizontal scaling.
	// The lock is re-entrant within the same transaction, so multiple events
	// of the same aggregate in one batch all pass. Locks release on commit or
	// rollback. FOR UPDATE SKIP LOCKED stays as a second guard against row
	// races during retries.
	const q = `
		SELECT id, aggregate_id, event_type, payload, created_at, trace_context
		  FROM outbox_events
		 WHERE published_at IS NULL
		   AND pg_try_advisory_xact_lock(hashtext(aggregate_id))
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
		if err := rows.Scan(&evt.ID, &evt.AggregateID, &evt.EventType, &payload, &evt.CreatedAt, &evt.TraceContext); err != nil {
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

func (r *postgresRepository) CountPending(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.QueryRowContext(ctx,
		`SELECT count(*) FROM outbox_events WHERE published_at IS NULL;`).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("outbox count pending: %w", err)
	}
	return n, nil
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
