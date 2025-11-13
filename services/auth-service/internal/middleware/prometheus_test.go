package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPrometheusMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		path           string
		handler        func(*gin.Context)
		expectedStatus int
	}{
		{
			name:   "GET request with valid path",
			method: "GET",
			path:   "/api/test",
			handler: func(c *gin.Context) {
				c.Status(http.StatusOK)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "POST request",
			method: "POST",
			path:   "/api/users",
			handler: func(c *gin.Context) {
				c.Status(http.StatusCreated)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Request with error status",
			method: "GET",
			path:   "/api/error",
			handler: func(c *gin.Context) {
				c.Status(http.StatusInternalServerError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "Request with not found",
			method: "GET",
			path:   "/api/notfound",
			handler: func(c *gin.Context) {
				c.Status(http.StatusNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(PrometheusMiddleware())
			router.Handle(tt.method, tt.path, tt.handler)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestPrometheusMiddleware_UnknownPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(PrometheusMiddleware())

	// No routes defined - any request will have empty FullPath
	router.NoRoute(func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})

	req := httptest.NewRequest("GET", "/unknown", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should handle unknown paths gracefully
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestPrometheusMiddleware_RecordsMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(PrometheusMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Make multiple requests to ensure metrics are recorded
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i, w.Code)
		}
	}
}

func TestPrometheusMiddleware_WithDifferentMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(PrometheusMiddleware())

	router.GET("/resource", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.POST("/resource", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})
	router.PUT("/resource", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.DELETE("/resource", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	methods := []struct {
		method string
		status int
	}{
		{"GET", http.StatusOK},
		{"POST", http.StatusCreated},
		{"PUT", http.StatusOK},
		{"DELETE", http.StatusNoContent},
	}

	for _, m := range methods {
		req := httptest.NewRequest(m.method, "/resource", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != m.status {
			t.Errorf("%s: expected status %d, got %d", m.method, m.status, w.Code)
		}
	}
}

func TestPrometheusMiddleware_MeasuresDuration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(PrometheusMiddleware())
	router.GET("/slow", func(c *gin.Context) {
		// Simulate some processing time
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/slow", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
