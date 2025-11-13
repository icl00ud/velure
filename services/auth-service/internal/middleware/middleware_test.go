package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkHeaders   bool
	}{
		{
			name:           "OPTIONS request returns 204",
			method:         "OPTIONS",
			expectedStatus: http.StatusNoContent,
			checkHeaders:   true,
		},
		{
			name:           "GET request passes through",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "POST request passes through",
			method:         "POST",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "PUT request passes through",
			method:         "PUT",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "DELETE request passes through",
			method:         "DELETE",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(CORS())
			router.Any("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkHeaders {
				// Verify CORS headers are set
				headers := map[string]string{
					"Access-Control-Allow-Origin":      "*",
					"Access-Control-Allow-Credentials": "true",
					"Access-Control-Allow-Methods":     "POST, OPTIONS, GET, PUT, DELETE",
				}

				for key, expected := range headers {
					actual := w.Header().Get(key)
					if actual != expected {
						t.Errorf("Header %s: expected '%s', got '%s'", key, expected, actual)
					}
				}

				// Check that Access-Control-Allow-Headers is set
				allowHeaders := w.Header().Get("Access-Control-Allow-Headers")
				if allowHeaders == "" {
					t.Error("Access-Control-Allow-Headers should be set")
				}
			}
		})
	}
}

func TestCORS_HeadersContent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify specific headers that should be allowed
	allowHeaders := w.Header().Get("Access-Control-Allow-Headers")
	expectedHeaders := []string{
		"Content-Type",
		"Authorization",
		"accept",
		"origin",
	}

	for _, header := range expectedHeaders {
		if !contains(allowHeaders, header) {
			t.Errorf("Access-Control-Allow-Headers should contain '%s', got '%s'", header, allowHeaders)
		}
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

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
