package middleware

import (
	"net/http"
	"time"

	"github.com/icl00ud/velure-shared/logger"
)

// skipLoggingPaths contains paths that should not be logged
var skipLoggingPaths = map[string]bool{
	"/metrics": true,
	"/health":  true,
	"/healthz": true,
	"/readyz":  true,
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(lrw, r)

		// Skip logging for health and metrics endpoints
		if skipLoggingPaths[r.URL.Path] {
			return
		}

		logger.Info("request",
			logger.String("method", r.Method),
			logger.String("path", r.URL.Path),
			logger.Int("status", lrw.statusCode),
			logger.Duration("duration", time.Since(start)),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Flush implements http.Flusher to support SSE
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
