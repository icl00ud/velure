package supervisor

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Budget limits how many restarts are allowed within a sliding window.
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
