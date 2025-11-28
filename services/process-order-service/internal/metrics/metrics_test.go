package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMetricsCounters(t *testing.T) {
	OrdersProcessed.WithLabelValues("success").Inc()
	PaymentAttempts.WithLabelValues("initiated").Inc()
	MessagesAcknowledged.WithLabelValues("ack").Inc()

	if got := testutil.ToFloat64(OrdersProcessed.WithLabelValues("success")); got != 1 {
		t.Fatalf("expected OrdersProcessed 1, got %f", got)
	}
	if got := testutil.ToFloat64(MessagesAcknowledged.WithLabelValues("ack")); got != 1 {
		t.Fatalf("expected MessagesAcknowledged 1, got %f", got)
	}
}
