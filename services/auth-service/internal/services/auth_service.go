package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"velure-auth-service/internal/config"
	"velure-auth-service/internal/metrics"
	"velure-auth-service/internal/models"
	"velure-auth-service/internal/repositories"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo          repositories.UserRepositoryInterface
	sessionRepo       repositories.SessionRepositoryInterface
	passwordResetRepo repositories.PasswordResetRepositoryInterface
	config            *config.Config
	bcryptWorkerPool  chan struct{}
	redis             *redis.Client
	tokenCache        sync.Map // fallback cache se redis falhar
}

func NewAuthService(
	userRepo repositories.UserRepositoryInterface,
	sessionRepo repositories.SessionRepositoryInterface,
	passwordResetRepo repositories.PasswordResetRepositoryInterface,
	config *config.Config,
	redisClient *redis.Client,
) *AuthService {
	// Worker pool limita operações bcrypt concorrentes (CPU-bound)
	workerPoolSize := config.Performance.BcryptWorkers
	if workerPoolSize <= 0 {
		workerPoolSize = 10 // default
	}
	bcryptWorkerPool := make(chan struct{}, workerPoolSize)

	return &AuthService{
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		passwordResetRepo: passwordResetRepo,
		config:            config,
		bcryptWorkerPool:  bcryptWorkerPool,
		redis:             redisClient,
	}
}

func (s *AuthService) CreateUser(req models.CreateUserRequest) (*models.RegistrationResponse, error) {
	start := time.Now()
	defer func() {
		metrics.RegistrationDuration.Observe(time.Since(start).Seconds())
	}()

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		metrics.RegistrationAttempts.WithLabelValues("failure").Inc()
		metrics.Errors.WithLabelValues("database").Inc()
		return nil, fmt.Errorf("error checking existing user: %w", err)
	}
	if existingUser != nil {
		metrics.RegistrationAttempts.WithLabelValues("conflict").Inc()
		return nil, errors.New("user already exists")
	}

	// Hash password com worker pool (evita sobrecarga CPU)
	s.bcryptWorkerPool <- struct{}{} // acquire worker
	hashedPassword, err := s.hashPasswordOptimized(req.Password)
	<-s.bcryptWorkerPool // release worker

	if err != nil {
		metrics.RegistrationAttempts.WithLabelValues("failure").Inc()
		metrics.Errors.WithLabelValues("internal").Inc()
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Create user
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(user); err != nil {
		metrics.RegistrationAttempts.WithLabelValues("failure").Inc()
		metrics.Errors.WithLabelValues("database").Inc()
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	// Criar session com tokens (versão síncrona otimizada)
	session, err := s.updateOrCreateSession(user.ID)
	if err != nil {
		metrics.RegistrationAttempts.WithLabelValues("failure").Inc()
		metrics.Errors.WithLabelValues("internal").Inc()
		return nil, fmt.Errorf("error creating session: %w", err)
	}

	// PERFORMANCE: Métricas de count removidas - devem ser coletadas por job periódico
	// para evitar sobrecarga de queries COUNT durante picos de carga
	// TODO: Implementar cronjob para atualizar total_users e active_sessions a cada 30s

	metrics.RegistrationAttempts.WithLabelValues("success").Inc()

	return &models.RegistrationResponse{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
	}, nil
}

func (s *AuthService) Login(req models.LoginRequest) (*models.LoginResponse, error) {
	start := time.Now()
	var status string
	defer func() {
		metrics.LoginDuration.WithLabelValues(status).Observe(time.Since(start).Seconds())
		metrics.LoginAttempts.WithLabelValues(status).Inc()
	}()

	// Get user by email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		status = "failure"
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		metrics.Errors.WithLabelValues("database").Inc()
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	// Check password com worker pool
	s.bcryptWorkerPool <- struct{}{} // acquire worker
	passwordErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	<-s.bcryptWorkerPool // release worker

	if passwordErr != nil {
		status = "failure"
		return nil, errors.New("invalid credentials")
	}

	// Criar session com tokens (versão síncrona otimizada)
	session, err := s.updateOrCreateSession(user.ID)
	if err != nil {
		status = "failure"
		metrics.Errors.WithLabelValues("internal").Inc()
		return nil, fmt.Errorf("error creating session: %w", err)
	}

	// PERFORMANCE: Métrica de count removida - deve ser coletada por job periódico
	// para evitar sobrecarga de queries COUNT durante picos de carga

	status = "success"
	return &models.LoginResponse{
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
	}, nil
}

func (s *AuthService) ValidateAccessToken(token string) (*models.User, error) {
	// Cache lookup (evita parsing e DB query repetidos)
	if s.config.Performance.EnableCache {
		if cachedUser, ok := s.tokenCache.Load(token); ok {
			if user, ok := cachedUser.(*models.User); ok {
				metrics.TokenValidations.WithLabelValues("valid_cached").Inc()
				return user, nil
			}
		}
	}

	claims := &jwt.RegisteredClaims{}

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil || !parsedToken.Valid {
		metrics.TokenValidations.WithLabelValues("invalid").Inc()
		return nil, errors.New("invalid token")
	}

	userID, err := strconv.ParseUint(claims.Subject, 10, 32)
	if err != nil {
		metrics.TokenValidations.WithLabelValues("invalid").Inc()
		return nil, errors.New("invalid user ID in token")
	}

	user, err := s.userRepo.GetByID(uint(userID))
	if err != nil {
		metrics.TokenValidations.WithLabelValues("invalid").Inc()
		return nil, errors.New("user not found")
	}

	// Cache token com TTL configurável
	if s.config.Performance.EnableCache {
		go func() {
			s.tokenCache.Store(token, user)
			cacheTTL := time.Duration(s.config.Performance.TokenCacheTTL) * time.Second
			time.AfterFunc(cacheTTL, func() {
				s.tokenCache.Delete(token)
			})
		}()
	}

	metrics.TokenValidations.WithLabelValues("valid").Inc()
	return user, nil
}

func (s *AuthService) GetUsers() ([]models.UserResponse, error) {
	metrics.UserQueries.WithLabelValues("list").Inc()

	users, err := s.userRepo.GetAll()
	if err != nil {
		metrics.Errors.WithLabelValues("database").Inc()
		return nil, fmt.Errorf("error getting users: %w", err)
	}

	responses := make([]models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	return responses, nil
}

func (s *AuthService) GetUsersByPage(page, pageSize int) (*models.PaginatedUsersResponse, error) {
	users, total, err := s.userRepo.GetByPage(page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("error getting users by page: %w", err)
	}

	responses := make([]models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	return models.NewPaginatedUsersResponse(responses, total, page, pageSize), nil
}

func (s *AuthService) GetUserByID(id uint) (*models.UserResponse, error) {
	metrics.UserQueries.WithLabelValues("by_id").Inc()

	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:id:%d", id)

	// Try Redis cache first
	if s.redis != nil {
		cachedJSON, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var cachedUser models.User
			if json.Unmarshal([]byte(cachedJSON), &cachedUser) == nil {
				metrics.CacheHits.Inc()
				response := cachedUser.ToResponse()
				return &response, nil
			}
		}
		metrics.CacheMisses.Inc()
	}

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		metrics.Errors.WithLabelValues("database").Inc()
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	// Cache in Redis
	if s.redis != nil {
		if userJSON, err := json.Marshal(user); err == nil {
			s.redis.Set(ctx, cacheKey, userJSON, time.Duration(s.config.Performance.TokenCacheTTL)*time.Second)
		}
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *AuthService) GetUserByEmail(email string) (*models.UserResponse, error) {
	metrics.UserQueries.WithLabelValues("by_email").Inc()

	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:email:%s", email)

	// Try Redis cache first
	if s.redis != nil {
		cachedJSON, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var cachedUser models.User
			if json.Unmarshal([]byte(cachedJSON), &cachedUser) == nil {
				metrics.CacheHits.Inc()
				response := cachedUser.ToResponse()
				return &response, nil
			}
		}
		metrics.CacheMisses.Inc()
	}

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		metrics.Errors.WithLabelValues("database").Inc()
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	// Cache in Redis
	if s.redis != nil {
		if userJSON, err := json.Marshal(user); err == nil {
			s.redis.Set(ctx, cacheKey, userJSON, time.Duration(s.config.Performance.TokenCacheTTL)*time.Second)
		}
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *AuthService) Logout(refreshToken string) error {
	metrics.LogoutRequests.Inc()

	if err := s.sessionRepo.InvalidateByRefreshToken(refreshToken); err != nil {
		metrics.Errors.WithLabelValues("database").Inc()
		return fmt.Errorf("error invalidating session: %w", err)
	}

	s.SyncActiveSessionsMetric(context.Background())
	return nil
}

func (s *AuthService) updateOrCreateSession(userID uint) (*models.Session, error) {
	// Generate tokens
	accessToken, err := s.generateAccessToken(userID)
	if err != nil {
		return nil, fmt.Errorf("error generating access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(userID)
	if err != nil {
		return nil, fmt.Errorf("error generating refresh token: %w", err)
	}

	// Check if session already exists
	existingSession, err := s.sessionRepo.GetByUserID(userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("error checking existing session: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(s.config.Session.ExpiresIn) * time.Millisecond)

	if existingSession != nil {
		// Update existing session
		existingSession.AccessToken = accessToken
		existingSession.RefreshToken = refreshToken
		existingSession.ExpiresAt = expiresAt

		if err := s.sessionRepo.Update(existingSession); err != nil {
			return nil, fmt.Errorf("error updating session: %w", err)
		}
		return existingSession, nil
	}

	// Create new session
	session := &models.Session{
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, fmt.Errorf("error creating session: %w", err)
	}

	return session, nil
}

func (s *AuthService) generateAccessToken(userID uint) (string, error) {
	start := time.Now()
	defer func() {
		metrics.TokenGenerationDuration.Observe(time.Since(start).Seconds())
	}()

	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(uint64(userID), 10),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.JWT.Secret))
	if err == nil {
		metrics.TokenGenerations.Inc()
	}
	return tokenString, err
}

func (s *AuthService) generateRefreshToken(userID uint) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(uint64(userID), 10),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	secret := s.config.JWT.Secret + s.config.JWT.RefreshSecret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// hashPasswordOptimized usa bcrypt com cost ajustável
func (s *AuthService) hashPasswordOptimized(password string) ([]byte, error) {
	// Cost 10 = ~100ms, Cost 12 = ~400ms, Cost 14 = ~1.6s
	// Para alta concorrência, usar cost menor (10) com worker pool
	cost := s.config.Performance.BcryptCost
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return bcrypt.GenerateFromPassword([]byte(password), cost)
}

// updateOrCreateSessionAsync versão async DEPRECATED - use updateOrCreateSession
// Mantido para compatibilidade, mas redireciona para versão síncrona (linha 294)
// PERFORMANCE: Async overhead removido, JWT signing é rápido (~1ms) e não justifica goroutines
func (s *AuthService) updateOrCreateSessionAsync(userID uint) (*models.Session, error) {
	return s.updateOrCreateSession(userID)
}

func (s *AuthService) SyncActiveSessionsMetric(ctx context.Context) {
	if s.sessionRepo == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	count, err := s.sessionRepo.CountActiveSessions(ctx)
	if err != nil {
		zap.L().Warn("failed to sync active sessions metric", zap.Error(err))
		return
	}
	metrics.ActiveSessions.Set(float64(count))
}

func (s *AuthService) SyncTotalUsersMetric(ctx context.Context) {
	if s.userRepo == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	count, err := s.userRepo.CountUsers(ctx)
	if err != nil {
		zap.L().Warn("failed to sync total users metric", zap.Error(err))
		return
	}
	metrics.TotalUsers.Set(float64(count))
}
