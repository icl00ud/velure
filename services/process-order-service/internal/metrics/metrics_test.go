package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMetricsIncrementAndObserve(t *testing.T) {
	OrdersProcessed.WithLabelValues("success").Inc()
	PaymentAttempts.WithLabelValues("success").Inc()
	InventoryChecks.WithLabelValues("available").Inc()
	MessagesAcknowledged.WithLabelValues("ack").Inc()
	ActiveWorkers.Set(4)
	OrderProcessingDuration.Observe(0.5)

	if got := testutil.ToFloat64(OrdersProcessed.WithLabelValues("success")); got != 1 {
		t.Fatalf("expected OrdersProcessed to be 1, got %f", got)
	}
	if got := testutil.ToFloat64(ActiveWorkers); got != 4 {
		t.Fatalf("expected ActiveWorkers to be 4, got %f", got)
	}
	if got := testutil.ToFloat64(MessagesAcknowledged.WithLabelValues("ack")); got != 1 {
		t.Fatalf("expected MessagesAcknowledged(ack) to be 1, got %f", got)
	}
}
