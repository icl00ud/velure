package middleware

import (
	"strconv"
	"time"

	"product-service/internal/metrics"

	"github.com/gofiber/fiber/v2"
)

// PrometheusMiddleware tracks HTTP request metrics for Fiber
func PrometheusMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip metrics endpoint to avoid recursion and errors
		if c.Path() == "/metrics" {
			return c.Next()
		}

		start := time.Now()

		// Process request
		err := c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()

		// Safely get the route path
		path := c.Path()
		route := c.Route()
		if route != nil && route.Path != "" {
			path = route.Path
		}
		if path == "" {
			path = "unknown"
		}

		status := strconv.Itoa(c.Response().StatusCode())

		metrics.HTTPRequests.WithLabelValues(
			c.Method(),
			path,
			status,
		).Inc()

		metrics.HTTPRequestDuration.WithLabelValues(
			c.Method(),
			path,
		).Observe(duration)

		return err
	}
}
