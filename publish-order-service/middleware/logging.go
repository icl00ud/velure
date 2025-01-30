package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware é um middleware para logar as requisições.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Utiliza um ResponseWriter personalizado para capturar o status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)
		duration := time.Since(start)
		log.Printf("[%s] %s %d %v", r.Method, r.URL.Path, lrw.statusCode, duration)
	})
}

// loggingResponseWriter captura o status code para logging.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captura o status code.
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
