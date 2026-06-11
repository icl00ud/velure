package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/icl00ud/velure/shared/logger"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for OPTIONS requests (CORS preflight)
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Authorization header first; httpOnly cookie as fallback (the
			// SPA authenticates via cookies set by the auth-service).
			var tokenString string
			if authHeader := r.Header.Get("Authorization"); authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) != 2 || parts[0] != "Bearer" {
					logger.Warn("invalid authorization header format")
					http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
					return
				}
				tokenString = parts[1]
			} else if cookie, err := r.Cookie("access_token"); err == nil {
				tokenString = cookie.Value
			}

			if tokenString == "" {
				logger.Warn("missing credentials (no authorization header or access_token cookie)")
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			claims := &jwt.RegisteredClaims{}

			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				logger.Warn("invalid token", logger.Err(err))
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			userID := claims.Subject
			if userID == "" {
				logger.Warn("missing user_id in token")
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}
