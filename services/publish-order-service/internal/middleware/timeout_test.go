package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimeout(t *testing.T) {
	tests := []struct {
		name           string
		timeout        time.Duration
		handlerDelay   time.Duration
		expectedStatus int
		expectTimeout  bool
	}{
		{
			name:           "request completes before timeout",
			timeout:        100 * time.Millisecond,
			handlerDelay:   10 * time.Millisecond,
			expectedStatus: http.StatusOK,
			expectTimeout:  false,
		},
		{
			name:           "request times out",
			timeout:        50 * time.Millisecond,
			handlerDelay:   200 * time.Millisecond,
			expectedStatus: http.StatusGatewayTimeout,
			expectTimeout:  true,
		},
		{
			name:           "request completes at boundary",
			timeout:        100 * time.Millisecond,
			handlerDelay:   90 * time.Millisecond,
			expectedStatus: http.StatusOK,
			expectTimeout:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(tt.handlerDelay)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"ok"}`))
			})

			timeoutMiddleware := Timeout(tt.timeout)
			wrappedHandler := timeoutMiddleware(handler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rr := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectTimeout {
				body := rr.Body.String()
				expectedBody := `{"error":"request timeout"}`
				if body != expectedBody {
					t.Errorf("expected body %s, got %s", expectedBody, body)
				}
			}
		})
	}
}

func TestTimeout_ContextCancellation(t *testing.T) {
	contextCancelled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		select {
		case <-r.Context().Done():
			contextCancelled = true
		default:
		}
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

	// Give the handler goroutine time to check context
	time.Sleep(250 * time.Millisecond)

	if !contextCancelled {
		t.Error("expected context to be cancelled in handler")
	}
}
