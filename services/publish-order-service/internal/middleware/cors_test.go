package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		origin         string
		allowedOrigins string
		expectedOrigin string
		expectCreds    string
		expectedStatus int
		shouldCallNext bool
	}{
		{
			name:           "OPTIONS request from allowed origin",
			method:         http.MethodOptions,
			origin:         "https://example.com",
			allowedOrigins: "https://example.com,https://shop.velure.local",
			expectedOrigin: "https://example.com",
			expectCreds:    "true",
			expectedStatus: http.StatusNoContent,
			shouldCallNext: false,
		},
		{
			name:           "GET request with allowed origin",
			method:         http.MethodGet,
			origin:         "https://example.com",
			allowedOrigins: "https://example.com",
			expectedOrigin: "https://example.com",
			expectCreds:    "true",
			expectedStatus: http.StatusOK,
			shouldCallNext: true,
		},
		{
			name:           "POST request without origin",
			method:         http.MethodPost,
			origin:         "",
			allowedOrigins: "https://example.com",
			expectedOrigin: "",
			expectCreds:    "",
			expectedStatus: http.StatusOK,
			shouldCallNext: true,
		},
		{
			name:           "DELETE request with disallowed origin",
			method:         http.MethodDelete,
			origin:         "http://localhost:3000",
			allowedOrigins: "https://example.com",
			expectedOrigin: "",
			expectCreds:    "",
			expectedStatus: http.StatusOK,
			shouldCallNext: true,
		},
		{
			name:           "GET request uses default allowlist",
			method:         http.MethodGet,
			origin:         "https://velure.local",
			allowedOrigins: "",
			expectedOrigin: "https://velure.local",
			expectCreds:    "true",
			expectedStatus: http.StatusOK,
			shouldCallNext: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.allowedOrigins != "" {
				t.Setenv("CORS_ALLOWED_ORIGINS", tt.allowedOrigins)
			} else {
				t.Setenv("CORS_ALLOWED_ORIGINS", "")
			}

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
			if allowCredentials != tt.expectCreds {
				t.Errorf("expected Access-Control-Allow-Credentials %q, got %q", tt.expectCreds, allowCredentials)
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
