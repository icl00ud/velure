package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMetricsCounters(t *testing.T) {
	OrdersCreated.WithLabelValues("success").Inc()
	OrdersPublished.WithLabelValues("failure").Inc()
	OrderStatusUpdates.WithLabelValues("old", "new").Inc()
	SSEMessagesSent.Inc()
	HTTPRequests.WithLabelValues("publish-order-service", "GET", "/health", "200").Inc()
	HTTPRequestDuration.WithLabelValues("publish-order-service", "GET", "/health").Observe(0.05)

	if got := testutil.ToFloat64(OrdersCreated.WithLabelValues("success")); got != 1 {
		t.Fatalf("expected OrdersCreated 1, got %f", got)
	}
	if got := testutil.ToFloat64(HTTPRequests.WithLabelValues("publish-order-service", "GET", "/health", "200")); got != 1 {
		t.Fatalf("expected HTTPRequests 1, got %f", got)
	}
}
