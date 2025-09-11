package config

import (
	"os"
	"strconv"
)

type Config struct {
	Environment string
	Port        string
	JWT         JWTConfig
	Session     SessionConfig
	Database    DatabaseConfig
}

type JWTConfig struct {
	Secret           string
	ExpiresIn        string
	RefreshSecret    string
	RefreshExpiresIn string
}

type SessionConfig struct {
	Secret    string
	ExpiresIn int64
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	URL      string
}

func Load() *Config {
	port, _ := strconv.Atoi(getEnv("POSTGRES_PORT", "5432"))
	sessionExpiresIn, _ := strconv.ParseInt(getEnv("SESSION_EXPIRES_IN", "86400000"), 10, 64)

	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("AUTH_SERVICE_APP_PORT", "3020"),
		JWT: JWTConfig{
			Secret:           getEnv("JWT_SECRET", "your-secret-key"),
			ExpiresIn:        getEnv("JWT_EXPIRES_IN", "1h"),
			RefreshSecret:    getEnv("JWT_REFRESH_TOKEN_SECRET", "your-refresh-secret"),
			RefreshExpiresIn: getEnv("JWT_REFRESH_TOKEN_EXPIRES_IN", "7d"),
		},
		Session: SessionConfig{
			Secret:    getEnv("SESSION_SECRET", "session-secret"),
			ExpiresIn: sessionExpiresIn,
		},
		Database: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     port,
			Username: getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", "password"),
			Database: getEnv("POSTGRES_DATABASE_NAME", "auth_db"),
			URL:      getEnv("POSTGRES_URL", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
