package outbox

import (
	"context"
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
		WithArgs("evt-1", "order-1", "order.created", []byte(`{"id":"order-1"}`), sqlmock.AnyArg(), sqlmock.AnyArg()).
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
	rows := sqlmock.NewRows([]string{"id", "aggregate_id", "event_type", "payload", "created_at", "trace_context"}).
		AddRow("evt-1", "order-1", "order.created", []byte(`{}`), now, "").
		AddRow("evt-2", "order-2", "order.created", []byte(`{}`), now, "")

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

var _ = sqlmock.NewResult // compile check that sqlmock is imported

// Multi-replica deployments must keep per-aggregate ordering: the fetch query
// claims whole aggregates via advisory xact locks so two relays never split
// events of the same order between them.
func TestFetchUnpublished_ClaimsWholeAggregates(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "aggregate_id", "event_type", "payload", "created_at", "trace_context"}).
		AddRow("evt-1", "order-1", "order.created", []byte(`{}`), now, "")

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM outbox_events .*pg_try_advisory_xact_lock\(hashtext\(aggregate_id\)\).*`).
		WithArgs(50).
		WillReturnRows(rows)

	repo := NewPostgresRepository(db)
	tx, events, err := repo.FetchUnpublished(context.Background(), 50)
	if err != nil { t.Fatalf("FetchUnpublished: %v", err) }
	defer tx.Rollback()

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}
