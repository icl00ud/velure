package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogging_CapturesStatus(t *testing.T) {
	handler := Logging(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	req := httptest.NewRequest("GET", "/foo", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTeapot {
		t.Fatalf("esperado 418; recebeu %d", w.Code)
	}
}

func TestLogging_SkipsHealthAndMetrics(t *testing.T) {
	handler := Logging(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	skipPaths := []string{"/metrics", "/health", "/healthz", "/readyz"}
	for _, path := range skipPaths {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
	// Just verifying no panic - the logger is internal to middleware
}

type flushRecorder struct {
	*httptest.ResponseRecorder
	flushed bool
}

func (f *flushRecorder) Flush() {
	f.flushed = true
}

func TestResponseWriterFlush_Delegates(t *testing.T) {
	rec := &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
	rw := &responseWriter{ResponseWriter: rec}

	rw.Flush()

	if !rec.flushed {
		t.Fatal("expected underlying flusher to be invoked")
	}
}
