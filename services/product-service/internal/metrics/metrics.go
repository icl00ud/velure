package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Product operation metrics
	ProductQueries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "product_queries_total",
			Help: "Total number of product queries",
		},
		[]string{"operation"}, // operation: get_all, get_by_name, get_by_page, get_by_category, get_count
	)

	ProductMutations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "product_mutations_total",
			Help: "Total number of product mutations",
		},
		[]string{"operation", "status"}, // operation: create, update, delete; status: success, failure
	)

	ProductOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "product_operation_duration_seconds",
			Help:    "Duration of product operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// Cache metrics
	CacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "product_cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	CacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "product_cache_misses_total",
			Help: "Total number of cache misses",
		},
	)

	CacheOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "product_cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "status"}, // operation: get, set, delete; status: success, failure
	)

	CacheOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "product_cache_operation_duration_seconds",
			Help:    "Duration of cache operations in seconds",
			Buckets: []float64{.0001, .0005, .001, .005, .01, .025, .05},
		},
		[]string{"operation"},
	)

	// Inventory metrics
	InventoryUpdates = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "product_inventory_updates_total",
			Help: "Total number of inventory updates",
		},
		[]string{"status"}, // status: success, failure, insufficient_stock
	)

	CurrentProductCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "product_catalog_total",
			Help: "Current total number of products in catalog",
		},
	)

	// Search metrics
	ProductSearches = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "product_searches_total",
			Help: "Total number of product searches",
		},
		[]string{"type"}, // type: by_name, by_category, paginated
	)

	SearchResultsReturned = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "product_search_results_count",
			Help:    "Number of results returned in search queries",
			Buckets: []float64{0, 1, 5, 10, 25, 50, 100, 500},
		},
	)

	// Database metrics
	MongoDBQueries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "product_mongodb_queries_total",
			Help: "Total number of MongoDB queries",
		},
		[]string{"operation", "collection"}, // operation: find, insert, update, delete
	)

	MongoDBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "product_mongodb_query_duration_seconds",
			Help:    "Duration of MongoDB queries in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5},
		},
		[]string{"operation"},
	)

	// Category metrics
	CategoryQueries = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "product_category_queries_total",
			Help: "Total number of category queries",
		},
	)

	// Error metrics
	Errors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "product_errors_total",
			Help: "Total number of errors",
		},
		[]string{"type"}, // type: validation, database, cache, not_found, internal
	)

	// HTTP metrics
	HTTPRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "product_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "product_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)
