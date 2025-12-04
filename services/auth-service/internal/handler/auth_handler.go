package handlers

import (
	"net/http"
	"strconv"
	"time"

	"velure-auth-service/internal/metrics"
	"velure-auth-service/internal/model"
	"velure-auth-service/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService services.AuthServiceInterface
}

func NewAuthHandler(authService services.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	start := time.Now()

	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		metrics.RegistrationAttempts.WithLabelValues("invalid_request").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Criar usuário (chamada direta, sem overhead de goroutine)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	metrics.RegistrationAttempts.WithLabelValues("success").Inc()
	metrics.RegistrationDuration.Observe(time.Since(start).Seconds())
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

	// Fazer login (chamada direta, sem overhead de goroutine)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	metrics.LoginAttempts.WithLabelValues("success").Inc()
	metrics.LoginDuration.WithLabelValues("success").Observe(time.Since(start).Seconds())
	metrics.TokenGenerations.Inc()
	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) ValidateToken(c *gin.Context) {
	var req models.ValidateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		metrics.TokenValidations.WithLabelValues("invalid_request").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	// Verifica se há parâmetros de paginação
	pageStr := c.Query("page")
	pageSizeStr := c.Query("pageSize")

	if pageStr != "" && pageSizeStr != "" {
		page, errPage := strconv.Atoi(pageStr)
		pageSize, errPageSize := strconv.Atoi(pageSizeStr)

		if errPage == nil && errPageSize == nil && page > 0 && pageSize > 0 {
			result, err := h.authService.GetUsersByPage(page, pageSize)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, result)
			return
		}
	}

	// Fallback para retornar todos os usuários
	users, err := h.authService.GetUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken := c.Param("refreshToken")
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh token is required"})
		return
	}

	if err := h.authService.Logout(refreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}
