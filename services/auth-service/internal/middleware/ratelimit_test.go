package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiterAllowRefill(t *testing.T) {
	limiter := newRateLimiterWithInterval(100, 2, 10*time.Millisecond)
	defer limiter.Stop()

	visitor := limiter.getVisitor("127.0.0.1")

	if !visitor.allow(limiter.rate, limiter.burst) {
		t.Fatalf("expected first request to be allowed")
	}
	if !visitor.allow(limiter.rate, limiter.burst) {
		t.Fatalf("expected second request to use burst token")
	}
	if visitor.allow(limiter.rate, limiter.burst) {
		t.Fatalf("expected third request to be blocked due to rate limit")
	}

	time.Sleep(20 * time.Millisecond)

	if !visitor.allow(limiter.rate, limiter.burst) {
		t.Fatalf("expected tokens to refill after interval")
	}
}

func TestRateLimiterMiddlewareBlocks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := newRateLimiterWithInterval(1, 1, time.Hour)
	defer limiter.Stop()

	router := gin.New()
	router.Use(limiter.Middleware())
	router.GET("/resource", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/resource", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected first request to succeed, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request to be rate limited, got %d", w.Code)
	}
}

func TestRateLimitByEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RateLimitByEndpoint(1, 1))
	router.GET("/a", func(c *gin.Context) {
		c.String(http.StatusOK, "a")
	})
	router.GET("/b", func(c *gin.Context) {
		c.String(http.StatusOK, "b")
	})

	reqA := httptest.NewRequest("GET", "/a", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, reqA)
	if w.Code != http.StatusOK {
		t.Fatalf("expected first request to /a to succeed, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	router.ServeHTTP(w, reqA)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request to /a to be limited, got %d", w.Code)
	}

	reqB := httptest.NewRequest("GET", "/b", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, reqB)
	if w.Code != http.StatusOK {
		t.Fatalf("expected request to different endpoint to pass, got %d", w.Code)
	}
}

func TestRateLimiterCleanupVisitors(t *testing.T) {
	limiter := newRateLimiterWithInterval(1, 1, 5*time.Millisecond)
	defer limiter.Stop()

	limiter.visitors["127.0.0.1"] = &Visitor{
		lastSeen: time.Now().Add(-4 * time.Minute),
		tokens:   0,
	}

	time.Sleep(15 * time.Millisecond)

	limiter.mu.RLock()
	_, exists := limiter.visitors["127.0.0.1"]
	limiter.mu.RUnlock()

	if exists {
		t.Fatalf("expected old visitor to be cleaned up")
	}
}
