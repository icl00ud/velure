package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimeout_PassesThroughResponseBeforeTimeout(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Test", "pass-through")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	timeoutMiddleware := Timeout(100 * time.Millisecond)
	wrappedHandler := timeoutMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	if rr.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected content type application/json, got %s", rr.Header().Get("Content-Type"))
	}

	if rr.Header().Get("X-Test") != "pass-through" {
		t.Fatalf("expected X-Test header to pass through, got %s", rr.Header().Get("X-Test"))
	}

	if rr.Body.String() != `{"status":"ok"}` {
		t.Fatalf("expected body %s, got %s", `{"status":"ok"}`, rr.Body.String())
	}
}

func TestTimeout_ReturnsDeterministicTimeoutResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"late"}`))
	})

	timeoutMiddleware := Timeout(50 * time.Millisecond)
	wrappedHandler := timeoutMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected status %d, got %d", http.StatusGatewayTimeout, rr.Code)
	}

	if rr.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected content type application/json, got %s", rr.Header().Get("Content-Type"))
	}

	expectedBody := `{"error":"request timeout"}`
	if rr.Body.String() != expectedBody {
		t.Fatalf("expected body %s, got %s", expectedBody, rr.Body.String())
	}
}

func TestTimeout_ContextCancellation(t *testing.T) {
	contextCancelled := make(chan struct{}, 1)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		contextCancelled <- struct{}{}
		w.WriteHeader(http.StatusOK)
	})

	timeoutMiddleware := Timeout(50 * time.Millisecond)
	wrappedHandler := timeoutMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusGatewayTimeout {
		t.Errorf("expected status %d, got %d", http.StatusGatewayTimeout, rr.Code)
	}

	select {
	case <-contextCancelled:
		return
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected context to be cancelled in handler")
	}
}
