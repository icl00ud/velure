package idempotency

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Checker dedupes redelivered events via Redis SET NX EX.
type Checker struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewChecker(rdb *redis.Client, ttl time.Duration) *Checker {
	return &Checker{rdb: rdb, ttl: ttl}
}

// FirstSeen reports whether eventID was seen for the first time.
// Internally issues `SET event:<id> 1 NX EX <ttl>` — atomic set+expire in one
// round-trip. Returns (true, nil) on first sight, (false, nil) on duplicate.
func (c *Checker) FirstSeen(ctx context.Context, eventID string) (bool, error) {
	return c.rdb.SetNX(ctx, "event:"+eventID, "1", c.ttl).Result()
}

// Forget removes the dedup record so a future delivery can reach the
// real handler. Called on handler failure to preserve retry semantics.
func (c *Checker) Forget(ctx context.Context, eventID string) error {
	return c.rdb.Del(ctx, "event:"+eventID).Err()
}
