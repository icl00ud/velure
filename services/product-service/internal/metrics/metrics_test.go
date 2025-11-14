package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestProductQueries(t *testing.T) {
	// Test that ProductQueries counter exists and can be incremented
	assert.NotNil(t, ProductQueries)

	// Test with different operations
	operations := []string{"get_all", "get_by_name", "get_by_page", "get_by_category", "get_count"}
	for _, op := range operations {
		ProductQueries.WithLabelValues(op).Inc()
	}

	// Verify metrics are registered
	metrics, err := prometheus.DefaultGatherer.Gather()
	assert.NoError(t, err)
	assert.NotEmpty(t, metrics)
}

func TestProductMutations(t *testing.T) {
	// Test that ProductMutations counter exists and can be incremented
	assert.NotNil(t, ProductMutations)

	// Test with different operations and statuses
	ProductMutations.WithLabelValues("create", "success").Inc()
	ProductMutations.WithLabelValues("create", "failure").Inc()
	ProductMutations.WithLabelValues("update", "success").Inc()
	ProductMutations.WithLabelValues("delete", "success").Inc()
	ProductMutations.WithLabelValues("delete", "failure").Inc()

	// Verify metrics are registered
	metrics, err := prometheus.DefaultGatherer.Gather()
	assert.NoError(t, err)
	assert.NotEmpty(t, metrics)
}

func TestProductOperationDuration(t *testing.T) {
	// Test that ProductOperationDuration histogram exists
	assert.NotNil(t, ProductOperationDuration)

	// Test observing durations
	operations := []string{"get_all", "get_by_name", "create", "update", "delete"}
	for _, op := range operations {
		ProductOperationDuration.WithLabelValues(op).Observe(0.123)
	}
}

func TestCacheMetrics(t *testing.T) {
	// Test cache counters
	assert.NotNil(t, CacheHits)
	assert.NotNil(t, CacheMisses)
	assert.NotNil(t, CacheOperations)
	assert.NotNil(t, CacheOperationDuration)

	// Increment cache hits and misses
	CacheHits.Inc()
	CacheMisses.Inc()

	// Test cache operations
	CacheOperations.WithLabelValues("get", "success").Inc()
	CacheOperations.WithLabelValues("set", "success").Inc()
	CacheOperations.WithLabelValues("delete", "success").Inc()
	CacheOperations.WithLabelValues("get", "failure").Inc()

	// Test cache operation duration
	CacheOperationDuration.WithLabelValues("get").Observe(0.001)
	CacheOperationDuration.WithLabelValues("set").Observe(0.002)
}

func TestInventoryMetrics(t *testing.T) {
	// Test inventory metrics
	assert.NotNil(t, InventoryUpdates)
	assert.NotNil(t, CurrentProductCount)

	// Test inventory updates with different statuses
	InventoryUpdates.WithLabelValues("success").Inc()
	InventoryUpdates.WithLabelValues("failure").Inc()
	InventoryUpdates.WithLabelValues("insufficient_stock").Inc()

	// Test current product count gauge
	CurrentProductCount.Set(100)
	CurrentProductCount.Set(150)
}

func TestSearchMetrics(t *testing.T) {
	// Test search metrics
	assert.NotNil(t, ProductSearches)
	assert.NotNil(t, SearchResultsReturned)

	// Test product searches
	ProductSearches.WithLabelValues("by_name").Inc()
	ProductSearches.WithLabelValues("by_category").Inc()
	ProductSearches.WithLabelValues("paginated").Inc()

	// Test search results histogram
	SearchResultsReturned.Observe(0)
	SearchResultsReturned.Observe(10)
	SearchResultsReturned.Observe(50)
	SearchResultsReturned.Observe(100)
}

func TestDatabaseMetrics(t *testing.T) {
	// Test database metrics
	assert.NotNil(t, MongoDBQueries)
	assert.NotNil(t, MongoDBQueryDuration)

	// Test MongoDB queries
	MongoDBQueries.WithLabelValues("find", "products").Inc()
	MongoDBQueries.WithLabelValues("insert", "products").Inc()
	MongoDBQueries.WithLabelValues("update", "products").Inc()
	MongoDBQueries.WithLabelValues("delete", "products").Inc()

	// Test query duration
	MongoDBQueryDuration.WithLabelValues("find").Observe(0.025)
	MongoDBQueryDuration.WithLabelValues("insert").Observe(0.015)
}

func TestCategoryMetrics(t *testing.T) {
	// Test category metrics
	assert.NotNil(t, CategoryQueries)

	// Increment category queries
	CategoryQueries.Inc()
	CategoryQueries.Inc()
}

func TestErrorMetrics(t *testing.T) {
	// Test error metrics
	assert.NotNil(t, Errors)

	// Test different error types
	Errors.WithLabelValues("validation").Inc()
	Errors.WithLabelValues("database").Inc()
	Errors.WithLabelValues("cache").Inc()
	Errors.WithLabelValues("not_found").Inc()
	Errors.WithLabelValues("internal").Inc()
}

func TestHTTPMetrics(t *testing.T) {
	// Test HTTP metrics
	assert.NotNil(t, HTTPRequests)
	assert.NotNil(t, HTTPRequestDuration)

	// Test HTTP requests
	HTTPRequests.WithLabelValues("product-service", "GET", "/products", "200").Inc()
	HTTPRequests.WithLabelValues("product-service", "POST", "/products", "201").Inc()
	HTTPRequests.WithLabelValues("product-service", "DELETE", "/products/:id", "204").Inc()
	HTTPRequests.WithLabelValues("product-service", "GET", "/products", "404").Inc()
	HTTPRequests.WithLabelValues("product-service", "GET", "/products", "500").Inc()

	// Test HTTP request duration
	HTTPRequestDuration.WithLabelValues("product-service", "GET", "/products").Observe(0.1)
	HTTPRequestDuration.WithLabelValues("product-service", "POST", "/products").Observe(0.2)
}

func TestAllMetricsInitialized(t *testing.T) {
	// Verify all metrics are properly initialized
	metricsToTest := []interface{}{
		ProductQueries,
		ProductMutations,
		ProductOperationDuration,
		CacheHits,
		CacheMisses,
		CacheOperations,
		CacheOperationDuration,
		InventoryUpdates,
		CurrentProductCount,
		ProductSearches,
		SearchResultsReturned,
		MongoDBQueries,
		MongoDBQueryDuration,
		CategoryQueries,
		Errors,
		HTTPRequests,
		HTTPRequestDuration,
	}

	for _, metric := range metricsToTest {
		assert.NotNil(t, metric)
	}
}

func TestMetricsRegistration(t *testing.T) {
	// Gather all metrics from default registry
	metrics, err := prometheus.DefaultGatherer.Gather()

	assert.NoError(t, err)
	assert.NotEmpty(t, metrics)

	// Check that we have various metric families registered
	metricNames := make(map[string]bool)
	for _, mf := range metrics {
		metricNames[mf.GetName()] = true
	}

	// Verify some of our custom metrics are registered
	expectedMetrics := []string{
		"product_queries_total",
		"product_mutations_total",
		"product_operation_duration_seconds",
		"product_cache_hits_total",
		"product_inventory_updates_total",
		"http_requests_total",
	}

	for _, name := range expectedMetrics {
		assert.True(t, metricNames[name], "Metric %s should be registered", name)
	}
}
