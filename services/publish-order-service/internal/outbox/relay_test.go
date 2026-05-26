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
	mu             sync.Mutex
	fetched        [][]model.OutboxEvent
	marked         [][]string
	fetchErr       error
	markErr        error
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
	mu       sync.Mutex
	calls    []model.OutboxEvent
	failOnID string
	err      error
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

func noopCommit(tx *sql.Tx) error   { return nil }
func noopRollback(tx *sql.Tx) error { return nil }
