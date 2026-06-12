package outbox

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/icl00ud/velure/services/publish-order-service/internal/metrics"
	"github.com/icl00ud/velure/services/publish-order-service/internal/model"
)

func TestCountPending(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM outbox_events WHERE published_at IS NULL")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))

	repo := NewPostgresRepository(db)
	n, err := repo.CountPending(context.Background())
	if err != nil {
		t.Fatalf("CountPending: %v", err)
	}
	if n != 7 {
		t.Fatalf("CountPending = %d, want 7", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRelay_ProcessBatch_UpdatesPendingGauge(t *testing.T) {
	repo := &fakeRepo{
		pendingBatches: [][]model.OutboxEvent{{
			{ID: "evt-1", EventType: "order.created", Payload: []byte(`{}`)},
		}},
		pendingCount: 3,
	}
	r := NewRelay(repo, &fakePublisher{}, WithCommitFn(noopCommit), WithRollbackFn(noopRollback))

	if err := r.processBatch(context.Background()); err != nil {
		t.Fatalf("processBatch: %v", err)
	}
	if got := testutil.ToFloat64(metrics.OutboxEventsPending); got != 3 {
		t.Fatalf("outbox_events_pending = %v, want 3", got)
	}
}

func TestRelay_ProcessBatch_UpdatesPendingGaugeOnPublishFailure(t *testing.T) {
	repo := &fakeRepo{
		pendingBatches: [][]model.OutboxEvent{{
			{ID: "evt-1", EventType: "order.created", Payload: []byte(`{}`)},
		}},
		pendingCount: 5,
	}
	pub := &fakePublisher{failOnID: "evt-1", err: errors.New("broker down")}
	r := NewRelay(repo, pub, WithCommitFn(noopCommit), WithRollbackFn(noopRollback))

	if err := r.processBatch(context.Background()); err == nil {
		t.Fatal("processBatch: want error, got nil")
	}
	if got := testutil.ToFloat64(metrics.OutboxEventsPending); got != 5 {
		t.Fatalf("outbox_events_pending = %v, want 5", got)
	}
}
