package testutil

import (
	"testing"
	"time"

	"velure-auth-service/internal/config"
	"velure-auth-service/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(&models.User{}, &models.Session{}, &models.PasswordReset{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	return db
}

// CreateTestUser creates a user with default values for testing
func CreateTestUser(overrides ...func(*models.User)) *models.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	user := &models.User{
		ID:        1,
		Name:      "Test User",
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, override := range overrides {
		override(user)
	}

	return user
}

// CreateTestUsers creates multiple users with different IDs and emails
func CreateTestUsers(count int) []*models.User {
	users := make([]*models.User, count)
	for i := 0; i < count; i++ {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		users[i] = &models.User{
			ID:        uint(i + 1),
			Name:      "Test User " + string(rune('A'+i)),
			Email:     "test" + string(rune('0'+i)) + "@example.com",
			Password:  string(hashedPassword),
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
			UpdatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
		}
	}
	return users
}

// CreateTestSession creates a session with default values for testing
func CreateTestSession(userID uint, overrides ...func(*models.Session)) *models.Session {
	session := &models.Session{
		ID:           1,
		UserID:       userID,
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:    time.Now(),
	}

	for _, override := range overrides {
		override(session)
	}

	return session
}

// CreateTestPasswordReset creates a password reset token with default values for testing
func CreateTestPasswordReset(userID uint, overrides ...func(*models.PasswordReset)) *models.PasswordReset {
	reset := &models.PasswordReset{
		ID:        1,
		UserID:    userID,
		Token:     "test-reset-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	for _, override := range overrides {
		override(reset)
	}

	return reset
}

// CreateTestCreateUserRequest creates a CreateUserRequest DTO for testing
func CreateTestCreateUserRequest(overrides ...func(*models.CreateUserRequest)) models.CreateUserRequest {
	req := models.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	for _, override := range overrides {
		override(&req)
	}

	return req
}

// CreateTestLoginRequest creates a LoginRequest DTO for testing
func CreateTestLoginRequest(overrides ...func(*models.LoginRequest)) models.LoginRequest {
	req := models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	for _, override := range overrides {
		override(&req)
	}

	return req
}

// CreateTestConfig creates a test configuration for JWT and sessions
func CreateTestConfig() *config.Config {
	return &config.Config{
		Environment: "test",
		Port:        "3020",
		JWT: config.JWTConfig{
			Secret:           "test-secret-key-for-testing-purposes-only",
			ExpiresIn:        "1h",
			RefreshSecret:    "test-refresh-secret-key-for-testing-purposes-only",
			RefreshExpiresIn: "7d",
		},
		Session: config.SessionConfig{
			Secret:    "test-session-secret",
			ExpiresIn: 604800, // 7 days in seconds
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Username: "test",
			Password: "test",
			Database: "test_db",
			URL:      "",
		},
	}
}

// HashPassword is a helper to hash passwords for testing
func HashPassword(password string) string {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword)
}

// ComparePasswords is a helper to compare passwords in tests
func ComparePasswords(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
