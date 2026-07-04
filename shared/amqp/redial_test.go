package amqp

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRedialSucceedsAfterRetries(t *testing.T) {
	attempts := 0
	var errs []error
	v, err := Redial(context.Background(), time.Millisecond, func(e error) {
		errs = append(errs, e)
	}, func(ctx context.Context) (int, error) {
		attempts++
		if attempts < 3 {
			return 0, errors.New("dial failed")
		}
		return 42, nil
	})
	if err != nil {
		t.Fatalf("expected success, got err %v", err)
	}
	if v != 42 {
		t.Fatalf("expected value 42, got %d", v)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
	if len(errs) != 2 {
		t.Fatalf("expected 2 onError calls, got %d", len(errs))
	}
}

func TestRedialStopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Redial(ctx, time.Millisecond, func(error) {}, func(ctx context.Context) (int, error) {
		return 0, errors.New("always fails")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
