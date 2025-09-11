package main

import (
	"testing"

	"velure-auth-service/internal/config"
	"velure-auth-service/internal/models"
	"velure-auth-service/internal/repositories"
	"velure-auth-service/internal/services"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Auto migrate
	err = db.AutoMigrate(&models.User{}, &models.Session{}, &models.PasswordReset{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestAuthService_CreateUser(t *testing.T) {
	db := setupTestDB(t)

	userRepo := repositories.NewUserRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	passwordResetRepo := repositories.NewPasswordResetRepository(db)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:           "test-secret",
			RefreshSecret:    "test-refresh-secret",
			ExpiresIn:        "1h",
			RefreshExpiresIn: "7d",
		},
		Session: config.SessionConfig{
			ExpiresIn: 86400000,
		},
	}

	authService := services.NewAuthService(userRepo, sessionRepo, passwordResetRepo, cfg)

	// Test creating a user
	req := models.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	user, err := authService.CreateUser(req)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, user.Email)
	}

	if user.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, user.Name)
	}

	// Test creating duplicate user
	_, err = authService.CreateUser(req)
	if err == nil {
		t.Error("Expected error when creating duplicate user")
	}
}

func TestAuthService_Login(t *testing.T) {
	db := setupTestDB(t)

	userRepo := repositories.NewUserRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	passwordResetRepo := repositories.NewPasswordResetRepository(db)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:           "test-secret",
			RefreshSecret:    "test-refresh-secret",
			ExpiresIn:        "1h",
			RefreshExpiresIn: "7d",
		},
		Session: config.SessionConfig{
			ExpiresIn: 86400000,
		},
	}

	authService := services.NewAuthService(userRepo, sessionRepo, passwordResetRepo, cfg)

	// Create a user first
	createReq := models.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	_, err := authService.CreateUser(createReq)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test login with correct credentials
	loginReq := models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	response, err := authService.Login(loginReq)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	if response.AccessToken == "" {
		t.Error("Expected access token to be generated")
	}

	if response.RefreshToken == "" {
		t.Error("Expected refresh token to be generated")
	}

	// Test login with incorrect password
	loginReq.Password = "wrongpassword"
	_, err = authService.Login(loginReq)
	if err == nil {
		t.Error("Expected error when logging in with wrong password")
	}
}
