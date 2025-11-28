package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMetricsIncrement(t *testing.T) {
	ProductQueries.WithLabelValues("get_all").Inc()
	ProductMutations.WithLabelValues("create", "success").Inc()
	CacheHits.Inc()
	CacheMisses.Add(2)
	CategoryQueries.Inc()
	HTTPRequests.WithLabelValues("GET", "/health", "200").Inc()
	HTTPRequestDuration.WithLabelValues("GET", "/health").Observe(0.1)

	if got := testutil.ToFloat64(ProductQueries.WithLabelValues("get_all")); got != 1 {
		t.Fatalf("expected ProductQueries counter 1, got %f", got)
	}
	if got := testutil.ToFloat64(CacheHits); got != 1 {
		t.Fatalf("expected CacheHits 1, got %f", got)
	}
	if got := testutil.ToFloat64(CacheMisses); got != 2 {
		t.Fatalf("expected CacheMisses 2, got %f", got)
	}
	if got := testutil.ToFloat64(HTTPRequests.WithLabelValues("GET", "/health", "200")); got != 1 {
		t.Fatalf("expected HTTPRequests counter 1, got %f", got)
	}
}
