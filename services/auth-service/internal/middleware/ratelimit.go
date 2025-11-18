package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	rate     int
	burst    int
}

type Visitor struct {
	lastSeen time.Time
	tokens   int
	mu       sync.Mutex
}

func NewRateLimiter(requestsPerSecond, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     requestsPerSecond,
		burst:    burst,
	}

	// Goroutine para limpar visitantes inativos
	go rl.cleanupVisitors()

	return rl
}

func (rl *RateLimiter) getVisitor(ip string) *Visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &Visitor{
			lastSeen: time.Now(),
			tokens:   rl.burst,
		}
		rl.visitors[ip] = v
	}
	return v
}

func (v *Visitor) allow(rate, burst int) bool {
	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(v.lastSeen).Seconds()
	v.lastSeen = now

	// Adicionar tokens baseado no tempo decorrido
	v.tokens += int(elapsed * float64(rate))
	if v.tokens > burst {
		v.tokens = burst
	}

	if v.tokens > 0 {
		v.tokens--
		return true
	}
	return false
}

func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			v.mu.Lock()
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
			v.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		visitor := rl.getVisitor(ip)

		// Verificar rate limit (visitor.allow já é thread-safe com mutex)
		if !visitor.allow(rl.rate, rl.burst) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByEndpoint aplica rate limiting específico por endpoint
func RateLimitByEndpoint(requestsPerSecond, burst int) gin.HandlerFunc {
	limiters := make(map[string]*RateLimiter)
	var mu sync.RWMutex

	return func(c *gin.Context) {
		key := c.ClientIP() + ":" + c.FullPath()

		mu.RLock()
		limiter, exists := limiters[key]
		mu.RUnlock()

		if !exists {
			mu.Lock()
			limiter = NewRateLimiter(requestsPerSecond, burst)
			limiters[key] = limiter
			mu.Unlock()
		}

		visitor := limiter.getVisitor(c.ClientIP())
		if !visitor.allow(requestsPerSecond, burst) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded for this endpoint",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
