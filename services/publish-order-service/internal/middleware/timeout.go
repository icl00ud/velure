package middleware

import (
	"bytes"
	"net/http"
	"time"
)

const (
	timeoutResponseBody     = `{"error":"request timeout"}`
	timeoutResponseSentinel = "__velure_timeout__"
)

type bufferedResponseWriter struct {
	headers    http.Header
	statusCode int
	body       bytes.Buffer
}

func newBufferedResponseWriter() *bufferedResponseWriter {
	return &bufferedResponseWriter{headers: make(http.Header)}
}

func (w *bufferedResponseWriter) Header() http.Header {
	return w.headers
}

func (w *bufferedResponseWriter) Write(data []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}

	return w.body.Write(data)
}

func (w *bufferedResponseWriter) WriteHeader(statusCode int) {
	if w.statusCode != 0 {
		return
	}

	w.statusCode = statusCode
}

func (w *bufferedResponseWriter) copyTo(dst http.ResponseWriter) {
	for key, values := range w.headers {
		for _, value := range values {
			dst.Header().Add(key, value)
		}
	}

	if w.statusCode != 0 {
		dst.WriteHeader(w.statusCode)
	}

	if w.body.Len() > 0 {
		_, _ = dst.Write(w.body.Bytes())
	}
}

func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		timeoutHandler := http.TimeoutHandler(next, duration, timeoutResponseSentinel)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bufferedWriter := newBufferedResponseWriter()
			timeoutHandler.ServeHTTP(bufferedWriter, r)

			if bufferedWriter.statusCode == http.StatusServiceUnavailable && bufferedWriter.body.String() == timeoutResponseSentinel {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusGatewayTimeout)
				_, _ = w.Write([]byte(timeoutResponseBody))
				return
			}

			bufferedWriter.copyTo(w)
		})
	}
}
