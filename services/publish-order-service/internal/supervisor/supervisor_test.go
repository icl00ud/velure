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
	if err == nil {
		t.Fatal("expected error after budget exhausted")
	}
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
