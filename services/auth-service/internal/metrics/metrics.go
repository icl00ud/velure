package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Login metrics
	LoginAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_login_attempts_total",
			Help: "Total number of login attempts",
		},
		[]string{"status"}, // status: success, failure
	)

	LoginDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_login_duration_seconds",
			Help:    "Duration of login requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)

	// Registration metrics
	RegistrationAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_registration_attempts_total",
			Help: "Total number of registration attempts",
		},
		[]string{"status"}, // status: success, failure, conflict
	)

	RegistrationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "auth_registration_duration_seconds",
			Help:    "Duration of registration requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	// Token metrics
	TokenValidations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_token_validations_total",
			Help: "Total number of token validation requests",
		},
		[]string{"result"}, // result: valid, invalid
	)

	TokenGenerations = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_token_generations_total",
			Help: "Total number of tokens generated",
		},
	)

	TokenGenerationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "auth_token_generation_duration_seconds",
			Help:    "Duration of token generation in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1},
		},
	)

	// Session metrics
	ActiveSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_active_sessions",
			Help: "Current number of active sessions",
		},
	)

	LogoutRequests = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_logout_requests_total",
			Help: "Total number of logout requests",
		},
	)

	// User metrics
	TotalUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_total_users",
			Help: "Total number of registered users",
		},
	)

	UserQueries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_user_queries_total",
			Help: "Total number of user queries",
		},
		[]string{"type"}, // type: by_id, by_email, list
	)

	// Database metrics
	DatabaseQueries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation"}, // operation: select, insert, update, delete
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_database_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25},
		},
		[]string{"operation"},
	)

	// Error metrics
	Errors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_errors_total",
			Help: "Total number of errors",
		},
		[]string{"type"}, // type: validation, database, auth, internal
	)

	// HTTP metrics
	HTTPRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Cache metrics
	CacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	CacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_cache_misses_total",
			Help: "Total number of cache misses",
		},
	)
)
