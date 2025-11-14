package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Order creation metrics
	OrdersCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "publish_order_created_total",
			Help: "Total number of orders created",
		},
		[]string{"status"}, // status: success, failure, validation_error
	)

	OrderCreationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "publish_order_creation_duration_seconds",
			Help:    "Duration of order creation in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	// Order publishing metrics
	OrdersPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "publish_order_published_total",
			Help: "Total number of orders published to RabbitMQ",
		},
		[]string{"status"}, // status: success, failure
	)

	RabbitMQPublishDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "publish_order_rabbitmq_publish_duration_seconds",
			Help:    "Duration of RabbitMQ publish operations in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25},
		},
	)

	// Order status metrics
	OrderStatusUpdates = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "publish_order_status_updates_total",
			Help: "Total number of order status updates",
		},
		[]string{"old_status", "new_status"},
	)

	CurrentOrdersByStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "publish_order_current_by_status",
			Help: "Current number of orders by status",
		},
		[]string{"status"}, // pending, processing, completed, failed
	)

	// SSE metrics
	SSEConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "publish_order_sse_connections",
			Help: "Current number of active SSE connections",
		},
	)

	SSEMessagesSent = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "publish_order_sse_messages_sent_total",
			Help: "Total number of SSE messages sent",
		},
	)

	// Pricing metrics
	OrderTotalValue = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "publish_order_total_value",
			Help:    "Total value of orders in currency units",
			Buckets: []float64{10, 50, 100, 250, 500, 1000, 2500, 5000, 10000},
		},
	)

	OrderItemsCount = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "publish_order_items_count",
			Help:    "Number of items in orders",
			Buckets: []float64{1, 2, 3, 5, 10, 15, 20, 30},
		},
	)

	// Database metrics
	DatabaseQueries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "publish_order_database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation"}, // operation: insert, update, select
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "publish_order_database_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25},
		},
		[]string{"operation"},
	)

	// Error metrics
	Errors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "publish_order_errors_total",
			Help: "Total number of errors",
		},
		[]string{"type"}, // type: validation, database, rabbitmq, pricing, internal
	)

	// HTTP metrics
	HTTPRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"service", "method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "path"},
	)
)
