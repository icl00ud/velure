package repository

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestClearProductCachesRemovesKeys(t *testing.T) {
	ctx := context.Background()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	repo := &productRepository{redis: rdb}

	// Seed cache keys the repository clears
	if err := rdb.Set(ctx, "allProducts", "value", 0).Err(); err != nil {
		t.Fatalf("set allProducts: %v", err)
	}
	if err := rdb.Set(ctx, "productCategories", "value", 0).Err(); err != nil {
		t.Fatalf("set productCategories: %v", err)
	}
	if err := rdb.Set(ctx, "productsPage:1:10", "value", 0).Err(); err != nil {
		t.Fatalf("set productsPage:1:10: %v", err)
	}
	if err := rdb.Set(ctx, "productsPage:2:10", "value", 0).Err(); err != nil {
		t.Fatalf("set productsPage:2:10: %v", err)
	}

	repo.clearProductCaches(ctx)

	keys, err := rdb.Keys(ctx, "*").Result()
	if err != nil {
		t.Fatalf("failed to list keys: %v", err)
	}
	if len(keys) != 0 {
		t.Fatalf("expected all cache keys to be cleared, remaining: %v", keys)
	}
}
