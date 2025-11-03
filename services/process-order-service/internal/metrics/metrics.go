package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Order processing metrics
	OrdersProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "process_order_processed_total",
			Help: "Total number of orders processed",
		},
		[]string{"status"}, // status: success, failure
	)

	OrderProcessingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "process_order_processing_duration_seconds",
			Help:    "Duration of order processing in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	// Payment metrics
	PaymentAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "process_order_payment_attempts_total",
			Help: "Total number of payment attempts",
		},
		[]string{"result"}, // result: success, failure, insufficient_funds
	)

	PaymentProcessingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "process_order_payment_processing_duration_seconds",
			Help:    "Duration of payment processing in seconds (simulation)",
			Buckets: []float64{.5, 1, 1.5, 2, 2.5, 3, 4, 5},
		},
	)

	PaymentTotalValue = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "process_order_payment_value",
			Help:    "Total value of payments in currency units",
			Buckets: []float64{10, 50, 100, 250, 500, 1000, 2500, 5000, 10000},
		},
	)

	// Inventory check metrics
	InventoryChecks = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "process_order_inventory_checks_total",
			Help: "Total number of inventory availability checks",
		},
		[]string{"result"}, // result: available, unavailable, error
	)

	InventoryCheckDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "process_order_inventory_check_duration_seconds",
			Help:    "Duration of inventory check API calls",
			Buckets: []float64{.01, .025, .05, .1, .25, .5, 1},
		},
	)

	// RabbitMQ consumer metrics
	MessagesConsumed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "process_order_messages_consumed_total",
			Help: "Total number of messages consumed from RabbitMQ",
		},
	)

	MessageProcessingErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "process_order_message_processing_errors_total",
			Help: "Total number of message processing errors",
		},
	)

	MessagesAcknowledged = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "process_order_messages_acknowledged_total",
			Help: "Total number of messages acknowledged",
		},
		[]string{"type"}, // type: ack, nack, reject
	)

	CurrentQueueSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "process_order_queue_size",
			Help: "Current estimated number of messages in queue (workers busy)",
		},
	)

	// Worker metrics
	ActiveWorkers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "process_order_active_workers",
			Help: "Current number of active worker goroutines",
		},
	)

	// Status update metrics
	StatusUpdatesPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "process_order_status_updates_published_total",
			Help: "Total number of status updates published back to RabbitMQ",
		},
		[]string{"status"}, // status: processing, completed, failed
	)

	// Product service client metrics
	ProductServiceCalls = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "process_order_product_service_calls_total",
			Help: "Total number of calls to product service",
		},
		[]string{"operation", "status"}, // operation: check_stock, update_stock; status: success, failure
	)

	ProductServiceCallDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "process_order_product_service_call_duration_seconds",
			Help:    "Duration of product service API calls",
			Buckets: []float64{.01, .025, .05, .1, .25, .5, 1, 2},
		},
		[]string{"operation"},
	)

	// Error metrics
	Errors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "process_order_errors_total",
			Help: "Total number of errors",
		},
		[]string{"type"}, // type: payment, inventory, product_service, rabbitmq, internal
	)
)
