package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/icl00ud/velure/services/auth-service/internal/config"
	"github.com/icl00ud/velure/services/auth-service/internal/metrics"
	"github.com/icl00ud/velure/services/auth-service/internal/model"
	"github.com/icl00ud/velure/services/auth-service/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/icl00ud/velure/shared/logger"
	"github.com/redis/go-redis/v9"
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
	tokenCache        sync.Map // fallback cache if redis fails
}

// userCacheEntry serializes/deserializes users in the Redis cache.
// Password is included here because models.User uses json:"-" to omit it.
type userCacheEntry struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewAuthService(
	userRepo repositories.UserRepositoryInterface,
	sessionRepo repositories.SessionRepositoryInterface,
	passwordResetRepo repositories.PasswordResetRepositoryInterface,
	config *config.Config,
	redisClient *redis.Client,
) *AuthService {
	// Worker pool caps concurrent bcrypt operations (CPU-bound)
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

	// Hash password through the worker pool (caps CPU pressure)
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

	// Create the token session (synchronous, optimized).
	session, err := s.updateOrCreateSession(user.ID)
	if err != nil {
		metrics.RegistrationAttempts.WithLabelValues("failure").Inc()
		metrics.Errors.WithLabelValues("internal").Inc()
		return nil, fmt.Errorf("error creating session: %w", err)
	}

	// Cache the freshly-registered user to speed up the first login.
	// Note: userCacheEntry is used to include the password in cache (json:"-" omits it from json.Marshal).
	s.cacheUser(context.Background(), user)

	// User/session count metrics are refreshed by the 30s background sync in main.

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

	ctx := context.Background()
	cacheKey := "user:email:" + req.Email

	// Try Redis cache first for user lookup
	var user *models.User
	fromCache := false
	if s.redis != nil {
		cachedJSON, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var cachedEntry userCacheEntry
			if json.Unmarshal([]byte(cachedJSON), &cachedEntry) == nil && cachedEntry.Password != "" {
				metrics.CacheHits.Inc()
				fromCache = true
				user = &models.User{
					ID:        cachedEntry.ID,
					Name:      cachedEntry.Name,
					Email:     cachedEntry.Email,
					Password:  cachedEntry.Password,
					CreatedAt: cachedEntry.CreatedAt,
					UpdatedAt: cachedEntry.UpdatedAt,
				}
			}
		}
		if user == nil {
			metrics.CacheMisses.Inc()
		}
	}

	// Get user by email from DB if not in cache
	if user == nil {
		var err error
		user, err = s.userRepo.GetByEmail(req.Email)
		if err != nil {
			status = "failure"
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("invalid credentials")
			}
			metrics.Errors.WithLabelValues("database").Inc()
			return nil, fmt.Errorf("error getting user: %w", err)
		}

		// Cache user in Redis for future logins
		s.cacheUser(ctx, user)
	}

	// Check password com worker pool
	passwordErr := s.comparePassword(user.Password, req.Password)

	// A cached hash can be stale (password changed after it was cached).
	// On mismatch, drop the cache entry, re-read the user from the DB and
	// compare again so a stale cache never locks the user out.
	if passwordErr != nil && fromCache {
		s.invalidateUserCache(ctx, user)

		freshUser, err := s.userRepo.GetByEmail(req.Email)
		if err != nil {
			status = "failure"
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("invalid credentials")
			}
			metrics.Errors.WithLabelValues("database").Inc()
			return nil, fmt.Errorf("error getting user: %w", err)
		}

		user = freshUser
		passwordErr = s.comparePassword(user.Password, req.Password)
		if passwordErr == nil {
			s.cacheUser(ctx, user)
		}
	}

	if passwordErr != nil {
		status = "failure"
		return nil, errors.New("invalid credentials")
	}

	// Create the token session (synchronous, optimized).
	session, err := s.updateOrCreateSession(user.ID)
	if err != nil {
		status = "failure"
		metrics.Errors.WithLabelValues("internal").Inc()
		return nil, fmt.Errorf("error creating session: %w", err)
	}

	// PERFORMANCE: count metric removed — should be gathered by a periodic job
	// to avoid hammering the DB with COUNT queries during traffic spikes.

	status = "success"
	return &models.LoginResponse{
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
	}, nil
}

func (s *AuthService) ValidateAccessToken(token string) (*models.User, error) {
	// Cache lookup (avoids repeated parsing and DB queries)
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
			var cachedEntry userCacheEntry
			if json.Unmarshal([]byte(cachedJSON), &cachedEntry) == nil {
				metrics.CacheHits.Inc()
				response := models.UserResponse{
					ID:        cachedEntry.ID,
					Name:      cachedEntry.Name,
					Email:     cachedEntry.Email,
					CreatedAt: cachedEntry.CreatedAt,
					UpdatedAt: cachedEntry.UpdatedAt,
				}
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
		cacheEntry := userCacheEntry{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Password:  user.Password,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
		if userJSON, err := json.Marshal(cacheEntry); err == nil {
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
			var cachedEntry userCacheEntry
			if json.Unmarshal([]byte(cachedJSON), &cachedEntry) == nil {
				metrics.CacheHits.Inc()
				response := models.UserResponse{
					ID:        cachedEntry.ID,
					Name:      cachedEntry.Name,
					Email:     cachedEntry.Email,
					CreatedAt: cachedEntry.CreatedAt,
					UpdatedAt: cachedEntry.UpdatedAt,
				}
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
		cacheEntry := userCacheEntry{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Password:  user.Password,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
		if userJSON, err := json.Marshal(cacheEntry); err == nil {
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

	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(uint64(userID), 10),
		ExpiresAt: jwt.NewNumericDate(now.Add(parseExpiry(s.config.JWT.ExpiresIn, time.Hour))),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.JWT.Secret))
	if err == nil {
		metrics.TokenGenerations.Inc()
	}
	return tokenString, err
}

func (s *AuthService) generateRefreshToken(userID uint) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(uint64(userID), 10),
		ExpiresAt: jwt.NewNumericDate(now.Add(parseExpiry(s.config.JWT.RefreshExpiresIn, 7*24*time.Hour))),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	secret := s.config.JWT.Secret + s.config.JWT.RefreshSecret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// parseExpiry parses durations like "1h", "30m" and the day shorthand "7d"
// (which time.ParseDuration does not support). Invalid or empty values fall
// back to the given default.
func parseExpiry(value string, fallback time.Duration) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	if strings.HasSuffix(value, "d") {
		if days, err := strconv.Atoi(strings.TrimSuffix(value, "d")); err == nil && days > 0 {
			return time.Duration(days) * 24 * time.Hour
		}
		return fallback
	}
	if d, err := time.ParseDuration(value); err == nil && d > 0 {
		return d
	}
	return fallback
}

// comparePassword runs the bcrypt comparison through the worker pool to bound
// concurrent hashing work.
func (s *AuthService) comparePassword(hash, password string) error {
	s.bcryptWorkerPool <- struct{}{} // acquire worker
	defer func() { <-s.bcryptWorkerPool }()
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// cacheUser stores the user (including its password hash, needed for login
// comparison) under both email and id keys.
func (s *AuthService) cacheUser(ctx context.Context, user *models.User) {
	if s.redis == nil {
		return
	}
	cacheEntry := userCacheEntry{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Password:  user.Password,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	userJSON, err := json.Marshal(cacheEntry)
	if err != nil {
		return
	}
	cacheTTL := time.Duration(s.config.Performance.TokenCacheTTL) * time.Second
	s.redis.Set(ctx, "user:email:"+user.Email, userJSON, cacheTTL)
	s.redis.Set(ctx, fmt.Sprintf("user:id:%d", user.ID), userJSON, cacheTTL)
}

// invalidateUserCache removes both cache keys for the user. Must be called by
// any flow that mutates user data (password change, profile update, delete).
func (s *AuthService) invalidateUserCache(ctx context.Context, user *models.User) {
	if s.redis == nil {
		return
	}
	s.redis.Del(ctx, "user:email:"+user.Email, fmt.Sprintf("user:id:%d", user.ID))
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
		logger.Warn("failed to sync active sessions metric", logger.Err(err))
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
		logger.Warn("failed to sync total users metric", logger.Err(err))
		return
	}
	metrics.TotalUsers.Set(float64(count))
}
