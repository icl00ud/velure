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
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return NewChecker(rdb, 24*time.Hour), mr
}

func TestFirstSeen_TrueOnNewKey(t *testing.T) {
	c, _ := newChecker(t)
	ok, err := c.FirstSeen(context.Background(), "evt-1")
	if err != nil || !ok {
		t.Fatalf("expected true, got ok=%v err=%v", ok, err)
	}
}

func TestFirstSeen_FalseOnDuplicate(t *testing.T) {
	c, _ := newChecker(t)
	if _, err := c.FirstSeen(context.Background(), "evt-1"); err != nil {
		t.Fatal(err)
	}
	ok, err := c.FirstSeen(context.Background(), "evt-1")
	if err != nil || ok {
		t.Fatalf("expected false, got ok=%v err=%v", ok, err)
	}
}

func TestForget_AllowsReprocessing(t *testing.T) {
	c, _ := newChecker(t)
	ctx := context.Background()
	_, _ = c.FirstSeen(ctx, "evt-1")
	if err := c.Forget(ctx, "evt-1"); err != nil {
		t.Fatal(err)
	}
	ok, err := c.FirstSeen(ctx, "evt-1")
	if err != nil || !ok {
		t.Fatalf("expected true after Forget, got %v %v", ok, err)
	}
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
