package config

import (
	"os"
	"testing"
)

func TestLoad_WithDefaultValues(t *testing.T) {
	// Clear environment variables
	os.Clearenv()

	cfg := Load()

	if cfg == nil {
		t.Fatal("Load() returned nil")
	}

	// Test default values
	if cfg.Environment != "development" {
		t.Errorf("Expected Environment 'development', got '%s'", cfg.Environment)
	}

	if cfg.Port != "3020" {
		t.Errorf("Expected Port '3020', got '%s'", cfg.Port)
	}

	if cfg.JWT.Secret != "your-secret-key" {
		t.Errorf("Expected JWT.Secret 'your-secret-key', got '%s'", cfg.JWT.Secret)
	}

	if cfg.JWT.ExpiresIn != "1h" {
		t.Errorf("Expected JWT.ExpiresIn '1h', got '%s'", cfg.JWT.ExpiresIn)
	}

	if cfg.JWT.RefreshSecret != "your-refresh-secret" {
		t.Errorf("Expected JWT.RefreshSecret 'your-refresh-secret', got '%s'", cfg.JWT.RefreshSecret)
	}

	if cfg.JWT.RefreshExpiresIn != "7d" {
		t.Errorf("Expected JWT.RefreshExpiresIn '7d', got '%s'", cfg.JWT.RefreshExpiresIn)
	}

	if cfg.Session.Secret != "session-secret" {
		t.Errorf("Expected Session.Secret 'session-secret', got '%s'", cfg.Session.Secret)
	}

	if cfg.Session.ExpiresIn != 86400000 {
		t.Errorf("Expected Session.ExpiresIn 86400000, got %d", cfg.Session.ExpiresIn)
	}

	if cfg.Database.Host != "localhost" {
		t.Errorf("Expected Database.Host 'localhost', got '%s'", cfg.Database.Host)
	}

	if cfg.Database.Port != 5432 {
		t.Errorf("Expected Database.Port 5432, got %d", cfg.Database.Port)
	}

	if cfg.Database.Username != "postgres" {
		t.Errorf("Expected Database.Username 'postgres', got '%s'", cfg.Database.Username)
	}

	if cfg.Database.Password != "password" {
		t.Errorf("Expected Database.Password 'password', got '%s'", cfg.Database.Password)
	}

	if cfg.Database.Database != "auth_db" {
		t.Errorf("Expected Database.Database 'auth_db', got '%s'", cfg.Database.Database)
	}
}

func TestLoad_WithCustomEnvironmentVariables(t *testing.T) {
	// Set custom environment variables
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("AUTH_SERVICE_APP_PORT", "8080")
	os.Setenv("JWT_SECRET", "custom-jwt-secret")
	os.Setenv("JWT_EXPIRES_IN", "2h")
	os.Setenv("JWT_REFRESH_TOKEN_SECRET", "custom-refresh-secret")
	os.Setenv("JWT_REFRESH_TOKEN_EXPIRES_IN", "30d")
	os.Setenv("SESSION_SECRET", "custom-session-secret")
	os.Setenv("SESSION_EXPIRES_IN", "604800000")
	os.Setenv("POSTGRES_HOST", "db.example.com")
	os.Setenv("POSTGRES_PORT", "5433")
	os.Setenv("POSTGRES_USER", "customuser")
	os.Setenv("POSTGRES_PASSWORD", "custompass")
	os.Setenv("POSTGRES_DATABASE_NAME", "custom_db")
	os.Setenv("POSTGRES_URL", "postgresql://custom:url")

	defer os.Clearenv()

	cfg := Load()

	if cfg == nil {
		t.Fatal("Load() returned nil")
	}

	// Test custom values
	if cfg.Environment != "production" {
		t.Errorf("Expected Environment 'production', got '%s'", cfg.Environment)
	}

	if cfg.Port != "8080" {
		t.Errorf("Expected Port '8080', got '%s'", cfg.Port)
	}

	if cfg.JWT.Secret != "custom-jwt-secret" {
		t.Errorf("Expected JWT.Secret 'custom-jwt-secret', got '%s'", cfg.JWT.Secret)
	}

	if cfg.JWT.ExpiresIn != "2h" {
		t.Errorf("Expected JWT.ExpiresIn '2h', got '%s'", cfg.JWT.ExpiresIn)
	}

	if cfg.JWT.RefreshSecret != "custom-refresh-secret" {
		t.Errorf("Expected JWT.RefreshSecret 'custom-refresh-secret', got '%s'", cfg.JWT.RefreshSecret)
	}

	if cfg.JWT.RefreshExpiresIn != "30d" {
		t.Errorf("Expected JWT.RefreshExpiresIn '30d', got '%s'", cfg.JWT.RefreshExpiresIn)
	}

	if cfg.Session.Secret != "custom-session-secret" {
		t.Errorf("Expected Session.Secret 'custom-session-secret', got '%s'", cfg.Session.Secret)
	}

	if cfg.Session.ExpiresIn != 604800000 {
		t.Errorf("Expected Session.ExpiresIn 604800000, got %d", cfg.Session.ExpiresIn)
	}

	if cfg.Database.Host != "db.example.com" {
		t.Errorf("Expected Database.Host 'db.example.com', got '%s'", cfg.Database.Host)
	}

	if cfg.Database.Port != 5433 {
		t.Errorf("Expected Database.Port 5433, got %d", cfg.Database.Port)
	}

	if cfg.Database.Username != "customuser" {
		t.Errorf("Expected Database.Username 'customuser', got '%s'", cfg.Database.Username)
	}

	if cfg.Database.Password != "custompass" {
		t.Errorf("Expected Database.Password 'custompass', got '%s'", cfg.Database.Password)
	}

	if cfg.Database.Database != "custom_db" {
		t.Errorf("Expected Database.Database 'custom_db', got '%s'", cfg.Database.Database)
	}

	if cfg.Database.URL != "postgresql://custom:url" {
		t.Errorf("Expected Database.URL 'postgresql://custom:url', got '%s'", cfg.Database.URL)
	}
}

func TestLoad_WithInvalidNumericValues(t *testing.T) {
	// Set invalid values for numeric fields
	os.Setenv("POSTGRES_PORT", "invalid")
	os.Setenv("SESSION_EXPIRES_IN", "not-a-number")

	defer os.Clearenv()

	cfg := Load()

	if cfg == nil {
		t.Fatal("Load() returned nil")
	}

	// When strconv.Atoi fails, it returns 0
	if cfg.Database.Port != 0 {
		t.Errorf("Expected Database.Port 0 (from invalid string), got %d", cfg.Database.Port)
	}

	if cfg.Session.ExpiresIn != 0 {
		t.Errorf("Expected Session.ExpiresIn 0 (from invalid string), got %d", cfg.Session.ExpiresIn)
	}
}
