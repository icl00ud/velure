//go:build integration

package main

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// These tests require `make local-up` to be running.
// Run: go test -tags=integration ./... -run Integration -v

func dsn() string {
	if v := os.Getenv("INTEGRATION_POSTGRES_DSN"); v != "" {
		return v
	}
	return "postgres://postgres:postgres@localhost:5432/publish_order_db?sslmode=disable"
}

func TestIntegration_OutboxHappyPath(t *testing.T) {
	// 1. POST /api/orders, get order id
	// 2. Poll outbox_events: assert row appears, published_at becomes non-null within 5s.
	t.Skip("requires JWT helper; flesh out when running")
}

func TestIntegration_RabbitDownThenUp(t *testing.T) {
	// 1. docker stop rabbitmq
	// 2. POST /api/orders
	// 3. Verify outbox row exists with published_at NULL
	// 4. docker start rabbitmq
	// 5. Verify published_at becomes non-null within 15s
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available")
	}
	t.Skip("flesh out with real docker controls")
}

func TestIntegration_DuplicateDelivery_IsSkipped(t *testing.T) {
	// 1. Publish the same event_id to the orders exchange twice via direct AMQP
	// 2. Query process_order_duplicates_skipped_total /metrics, assert it incremented by 1
	t.Skip("flesh out with direct AMQP publish")
}

func pollUntil(ctx context.Context, t *testing.T, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(200 * time.Millisecond):
		}
	}
	t.Fatal("condition not met within deadline")
}

// Suppress unused warnings for skeletons:
var _ *sql.DB = nil
