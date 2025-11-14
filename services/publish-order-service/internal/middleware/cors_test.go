package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	tests := []struct {
		name                  string
		method                string
		origin                string
		expectedOrigin        string
		expectedStatus        int
		shouldCallNext        bool
	}{
		{
			name:                  "OPTIONS request",
			method:                http.MethodOptions,
			origin:                "https://example.com",
			expectedOrigin:        "https://example.com",
			expectedStatus:        http.StatusNoContent,
			shouldCallNext:        false,
		},
		{
			name:                  "GET request with origin",
			method:                http.MethodGet,
			origin:                "https://example.com",
			expectedOrigin:        "https://example.com",
			expectedStatus:        http.StatusOK,
			shouldCallNext:        true,
		},
		{
			name:                  "POST request without origin",
			method:                http.MethodPost,
			origin:                "",
			expectedOrigin:        "*",
			expectedStatus:        http.StatusOK,
			shouldCallNext:        true,
		},
		{
			name:                  "DELETE request with origin",
			method:                http.MethodDelete,
			origin:                "http://localhost:3000",
			expectedOrigin:        "http://localhost:3000",
			expectedStatus:        http.StatusOK,
			shouldCallNext:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			corsHandler := CORS(handler)

			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			rr := httptest.NewRecorder()
			corsHandler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check CORS headers
			allowOrigin := rr.Header().Get("Access-Control-Allow-Origin")
			if allowOrigin != tt.expectedOrigin {
				t.Errorf("expected Access-Control-Allow-Origin %s, got %s", tt.expectedOrigin, allowOrigin)
			}

			allowMethods := rr.Header().Get("Access-Control-Allow-Methods")
			if allowMethods == "" {
				t.Error("expected Access-Control-Allow-Methods header to be set")
			}

			allowHeaders := rr.Header().Get("Access-Control-Allow-Headers")
			if allowHeaders == "" {
				t.Error("expected Access-Control-Allow-Headers header to be set")
			}

			allowCredentials := rr.Header().Get("Access-Control-Allow-Credentials")
			if allowCredentials != "true" {
				t.Errorf("expected Access-Control-Allow-Credentials to be true, got %s", allowCredentials)
			}

			maxAge := rr.Header().Get("Access-Control-Max-Age")
			if maxAge != "86400" {
				t.Errorf("expected Access-Control-Max-Age to be 86400, got %s", maxAge)
			}

			// Check if next handler was called
			if nextCalled != tt.shouldCallNext {
				t.Errorf("expected next handler called: %v, got: %v", tt.shouldCallNext, nextCalled)
			}
		})
	}
}
