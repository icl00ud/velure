package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAuth(t *testing.T) {
	secret := "test-secret"

	// Create a valid token
	validToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "user123",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})
	validTokenString, _ := validToken.SignedString([]byte(secret))

	// Create an expired token
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "user123",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
	})
	expiredTokenString, _ := expiredToken.SignedString([]byte(secret))

	// Create a token with wrong signature
	wrongSecretToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "user123",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})
	wrongSecretTokenString, _ := wrongSecretToken.SignedString([]byte("wrong-secret"))

	tests := []struct {
		name           string
		authHeader     string
		method         string
		expectedStatus int
		expectUserID   bool
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + validTokenString,
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectUserID:   true,
		},
		{
			name:           "options request - skip auth",
			authHeader:     "",
			method:         http.MethodOptions,
			expectedStatus: http.StatusOK,
			expectUserID:   false,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			method:         http.MethodGet,
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
		{
			name:           "invalid authorization header format - no bearer",
			authHeader:     validTokenString,
			method:         http.MethodGet,
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
		{
			name:           "invalid authorization header format - wrong prefix",
			authHeader:     "Token " + validTokenString,
			method:         http.MethodGet,
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
		{
			name:           "expired token",
			authHeader:     "Bearer " + expiredTokenString,
			method:         http.MethodGet,
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
		{
			name:           "token with wrong signature",
			authHeader:     "Bearer " + wrongSecretTokenString,
			method:         http.MethodGet,
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
		{
			name:           "invalid token format",
			authHeader:     "Bearer invalid.token.here",
			method:         http.MethodGet,
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that checks if userID is in context
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID := GetUserID(r.Context())
				if tt.expectUserID && userID == "" {
					t.Error("expected userID in context, got empty string")
				}
				if !tt.expectUserID && userID != "" {
					t.Errorf("expected no userID in context, got %s", userID)
				}
				w.WriteHeader(http.StatusOK)
			})

			// Wrap with Auth middleware
			authMiddleware := Auth(secret)
			wrappedHandler := authMiddleware(handler)

			// Create test request
			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Record response
			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestAuth_TokenWithoutSubject(t *testing.T) {
	secret := "test-secret"

	// Create a token without subject
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})
	tokenString, _ := token.SignedString([]byte(secret))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authMiddleware := Auth(secret)
	wrappedHandler := authMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(r *http.Request) *http.Request
		expected string
	}{
		{
			name: "context with valid user_id",
			setup: func(r *http.Request) *http.Request {
				ctx := r.Context()
				ctx = context.WithValue(ctx, UserIDKey, "user123")
				return r.WithContext(ctx)
			},
			expected: "user123",
		},
		{
			name: "context without user_id",
			setup: func(r *http.Request) *http.Request {
				return r
			},
			expected: "",
		},
		{
			name: "context with wrong type",
			setup: func(r *http.Request) *http.Request {
				ctx := r.Context()
				ctx = context.WithValue(ctx, UserIDKey, 12345) // wrong type
				return r.WithContext(ctx)
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = tt.setup(req)

			userID := GetUserID(req.Context())
			if userID != tt.expected {
				t.Errorf("expected userID %s, got %s", tt.expected, userID)
			}
		})
	}
}
