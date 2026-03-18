package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com,https://shop.velure.local")

	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Fatalf("expected Access-Control-Allow-Origin to echo allowed origin, got %q", got)
	}

	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("expected Access-Control-Allow-Credentials true, got %q", got)
	}

	if got := w.Header().Get("Access-Control-Allow-Methods"); got != "POST, OPTIONS, GET, PUT, DELETE" {
		t.Fatalf("unexpected Access-Control-Allow-Methods: %q", got)
	}

	allowHeaders := w.Header().Get("Access-Control-Allow-Headers")
	if allowHeaders == "" {
		t.Fatal("expected Access-Control-Allow-Headers to be set")
	}

	for _, header := range []string{"Content-Type", "Authorization", "origin"} {
		if !strings.Contains(allowHeaders, header) {
			t.Fatalf("expected Access-Control-Allow-Headers to contain %q, got %q", header, allowHeaders)
		}
	}
}

func TestCORS_DefaultAllowlist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://velure.local")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://velure.local" {
		t.Fatalf("expected default allowed origin to be echoed, got %q", got)
	}
}

func TestCORS_DisallowedOriginDoesNotSetAllowOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com")

	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected Access-Control-Allow-Origin to be empty for disallowed origin, got %q", got)
	}

	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("expected Access-Control-Allow-Credentials to be empty for disallowed origin, got %q", got)
	}
}

func TestCORS_MissingOriginDoesNotSetAllowOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com")

	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected Access-Control-Allow-Origin to be empty when Origin is missing, got %q", got)
	}
}

func TestCORS_OPTIONSReturnsNoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com")

	router := gin.New()
	router.Use(CORS())
	router.Any("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Fatalf("expected Access-Control-Allow-Origin to echo allowed origin, got %q", got)
	}
}

func TestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestLogger_WithMultipleRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	router.POST("/test", func(c *gin.Context) {
		c.String(http.StatusCreated, "Created")
	})

	tests := []struct {
		method         string
		expectedStatus int
	}{
		{"GET", http.StatusOK},
		{"POST", http.StatusCreated},
		{"GET", http.StatusOK},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != tt.expectedStatus {
			t.Errorf("Method %s: expected status %d, got %d", tt.method, tt.expectedStatus, w.Code)
		}
	}
}
