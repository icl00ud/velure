package handlers

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/icl00ud/velure/services/auth-service/internal/metrics"
	"github.com/icl00ud/velure/services/auth-service/internal/model"
	"github.com/icl00ud/velure/services/auth-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/icl00ud/velure/shared/logger"
)

type AuthHandler struct {
	authService services.AuthServiceInterface
}

func NewAuthHandler(authService services.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Auth cookies: tokens also travel as httpOnly cookies so the SPA never has
// to persist them in localStorage (XSS cannot read httpOnly cookies). The
// JSON body still includes them for API clients.
const (
	accessTokenCookie  = "access_token"
	refreshTokenCookie = "refresh_token"
)

func cookieSecure() bool {
	return os.Getenv("ENVIRONMENT") == "production"
}

// cookieMaxAge parses durations like "1h" or "7d"; invalid values fall back.
func cookieMaxAge(envKey string, fallback time.Duration) int {
	v := strings.TrimSpace(os.Getenv(envKey))
	if v == "" {
		return int(fallback.Seconds())
	}
	if strings.HasSuffix(v, "d") {
		if days, err := strconv.Atoi(strings.TrimSuffix(v, "d")); err == nil && days > 0 {
			return days * 24 * 3600
		}
		return int(fallback.Seconds())
	}
	if d, err := time.ParseDuration(v); err == nil && d > 0 {
		return int(d.Seconds())
	}
	return int(fallback.Seconds())
}

func setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(accessTokenCookie, accessToken,
		cookieMaxAge("JWT_EXPIRES_IN", time.Hour), "/", "", cookieSecure(), true)
	c.SetCookie(refreshTokenCookie, refreshToken,
		cookieMaxAge("JWT_REFRESH_TOKEN_EXPIRES_IN", 7*24*time.Hour), "/", "", cookieSecure(), true)
}

func clearAuthCookies(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(accessTokenCookie, "", -1, "/", "", cookieSecure(), true)
	c.SetCookie(refreshTokenCookie, "", -1, "/", "", cookieSecure(), true)
}

// internalError logs the real cause and returns a generic 500 so database and
// infrastructure details never reach the client.
func internalError(c *gin.Context, err error) {
	logger.Error("internal error", logger.Err(err))
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
}

func (h *AuthHandler) Register(c *gin.Context) {
	start := time.Now()

	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		metrics.RegistrationAttempts.WithLabelValues("invalid_request").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.CreateUser(req)
	if err != nil {
		if err.Error() == "user already exists" {
			metrics.RegistrationAttempts.WithLabelValues("conflict").Inc()
			metrics.RegistrationDuration.Observe(time.Since(start).Seconds())
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		metrics.RegistrationAttempts.WithLabelValues("failure").Inc()
		metrics.RegistrationDuration.Observe(time.Since(start).Seconds())
		internalError(c, err)
		return
	}

	metrics.RegistrationAttempts.WithLabelValues("success").Inc()
	metrics.RegistrationDuration.Observe(time.Since(start).Seconds())
	setAuthCookies(c, user.AccessToken, user.RefreshToken)
	c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	start := time.Now()

	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		metrics.LoginAttempts.WithLabelValues("invalid_request").Inc()
		metrics.LoginDuration.WithLabelValues("failure").Observe(time.Since(start).Seconds())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.authService.Login(req)
	if err != nil {
		if err.Error() == "invalid credentials" {
			metrics.LoginAttempts.WithLabelValues("invalid_credentials").Inc()
			metrics.LoginDuration.WithLabelValues("failure").Observe(time.Since(start).Seconds())
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		metrics.LoginAttempts.WithLabelValues("failure").Inc()
		metrics.LoginDuration.WithLabelValues("failure").Observe(time.Since(start).Seconds())
		internalError(c, err)
		return
	}

	metrics.LoginAttempts.WithLabelValues("success").Inc()
	metrics.LoginDuration.WithLabelValues("success").Observe(time.Since(start).Seconds())
	metrics.TokenGenerations.Inc()
	setAuthCookies(c, response.AccessToken, response.RefreshToken)
	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) ValidateToken(c *gin.Context) {
	var req models.ValidateTokenRequest
	_ = c.ShouldBindJSON(&req) // body optional: cookie fallback below

	if req.AccessToken == "" {
		if cookie, err := c.Cookie(accessTokenCookie); err == nil {
			req.AccessToken = cookie
		}
	}
	if req.AccessToken == "" {
		metrics.TokenValidations.WithLabelValues("invalid_request").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	user, err := h.authService.ValidateAccessToken(req.AccessToken)
	isValid := err == nil && user != nil

	if isValid {
		metrics.TokenValidations.WithLabelValues("valid").Inc()
	} else {
		metrics.TokenValidations.WithLabelValues("invalid").Inc()
	}

	c.JSON(http.StatusOK, models.ValidateTokenResponse{IsValid: isValid})
}

func (h *AuthHandler) GetUsers(c *gin.Context) {
	email := c.Query("email")
	if email != "" {
		user, err := h.authService.GetUserByEmail(email)
		if err != nil {
			if err.Error() == "user not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			internalError(c, err)
			return
		}

		c.JSON(http.StatusOK, user)
		return
	}

	pageStr := c.Query("page")
	pageSizeStr := c.Query("pageSize")

	if pageStr != "" && pageSizeStr != "" {
		page, errPage := strconv.Atoi(pageStr)
		pageSize, errPageSize := strconv.Atoi(pageSizeStr)

		if errPage == nil && errPageSize == nil && page > 0 && pageSize > 0 {
			result, err := h.authService.GetUsersByPage(page, pageSize)
			if err != nil {
				internalError(c, err)
				return
			}
			c.JSON(http.StatusOK, result)
			return
		}
	}

	users, err := h.authService.GetUsers()
	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *AuthHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	user, err := h.authService.GetUserByID(uint(id))
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	user, err := h.authService.GetUserByEmail(email)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req models.LogoutRequest
	_ = c.ShouldBindJSON(&req) // body optional: cookie fallback below

	if req.RefreshToken == "" {
		if cookie, err := c.Cookie(refreshTokenCookie); err == nil {
			req.RefreshToken = cookie
		}
	}
	if req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh token is required"})
		return
	}

	if err := h.authService.Logout(req.RefreshToken); err != nil {
		internalError(c, err)
		return
	}

	clearAuthCookies(c)
	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}
