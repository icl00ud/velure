package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	allowedOrigins := loadAllowedOrigins()

	return gin.HandlerFunc(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowedOrigins[origin]; origin != "" && ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
}

func loadAllowedOrigins() map[string]struct{} {
	originsConfig := os.Getenv("CORS_ALLOWED_ORIGINS")
	if strings.TrimSpace(originsConfig) == "" {
		originsConfig = "https://velure.local"
	}

	allowedOrigins := make(map[string]struct{})
	for _, origin := range strings.Split(originsConfig, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		allowedOrigins[trimmed] = struct{}{}
	}

	return allowedOrigins
}

func Logger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/metrics", "/health"},
	})
}
