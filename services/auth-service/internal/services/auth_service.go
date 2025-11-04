package services

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"velure-auth-service/internal/config"
	"velure-auth-service/internal/metrics"
	"velure-auth-service/internal/models"
	"velure-auth-service/internal/repositories"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo          *repositories.UserRepository
	sessionRepo       *repositories.SessionRepository
	passwordResetRepo *repositories.PasswordResetRepository
	config            *config.Config
}

func NewAuthService(
	userRepo *repositories.UserRepository,
	sessionRepo *repositories.SessionRepository,
	passwordResetRepo *repositories.PasswordResetRepository,
	config *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		passwordResetRepo: passwordResetRepo,
		config:            config,
	}
}

func (s *AuthService) CreateUser(req models.CreateUserRequest) (*models.UserResponse, error) {
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

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
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

	metrics.RegistrationAttempts.WithLabelValues("success").Inc()
	metrics.TotalUsers.Inc()
	response := user.ToResponse()
	return &response, nil
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

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		status = "failure"
		return nil, errors.New("invalid credentials")
	}

	// Create or update session
	session, err := s.updateOrCreateSession(user.ID)
	if err != nil {
		status = "failure"
		metrics.Errors.WithLabelValues("internal").Inc()
		return nil, fmt.Errorf("error creating session: %w", err)
	}

	status = "success"
	return &models.LoginResponse{
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
	}, nil
}

func (s *AuthService) ValidateAccessToken(token string) (*models.User, error) {
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

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		metrics.Errors.WithLabelValues("database").Inc()
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *AuthService) GetUserByEmail(email string) (*models.UserResponse, error) {
	metrics.UserQueries.WithLabelValues("by_email").Inc()

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		metrics.Errors.WithLabelValues("database").Inc()
		return nil, fmt.Errorf("error getting user: %w", err)
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
