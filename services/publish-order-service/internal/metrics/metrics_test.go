package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMetricsAreRecorded(t *testing.T) {
	OrdersCreated.WithLabelValues("success").Inc()
	OrdersPublished.WithLabelValues("failure").Inc()
	CurrentOrdersByStatus.WithLabelValues("pending").Set(3)
	SSEConnections.Set(2)
	SSEMessagesSent.Inc()
	DatabaseQueries.WithLabelValues("insert").Inc()
	HTTPRequests.WithLabelValues("publish-order-service", "GET", "/health", "200").Inc()

	if got := testutil.ToFloat64(OrdersCreated.WithLabelValues("success")); got != 1 {
		t.Fatalf("expected OrdersCreated counter to be 1, got %f", got)
	}
	if got := testutil.ToFloat64(CurrentOrdersByStatus.WithLabelValues("pending")); got != 3 {
		t.Fatalf("expected CurrentOrdersByStatus gauge to be 3, got %f", got)
	}
	if got := testutil.ToFloat64(HTTPRequests.WithLabelValues("publish-order-service", "GET", "/health", "200")); got != 1 {
		t.Fatalf("expected HTTPRequests counter to be 1, got %f", got)
	}
}
