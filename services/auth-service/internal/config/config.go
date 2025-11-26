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
	Redis       RedisConfig
	Performance PerformanceConfig
}

type PerformanceConfig struct {
	BcryptCost    int
	BcryptWorkers int
	TokenCacheTTL int // seconds
	EnableCache   bool
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

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func Load() *Config {
	port, _ := strconv.Atoi(getEnv("POSTGRES_PORT", "5432"))
	sessionExpiresIn, _ := strconv.ParseInt(getEnv("SESSION_EXPIRES_IN", "86400000"), 10, 64)
	bcryptCost, _ := strconv.Atoi(getEnv("BCRYPT_COST", "10"))
	bcryptWorkers, _ := strconv.Atoi(getEnv("BCRYPT_WORKERS", "10"))
	tokenCacheTTL, _ := strconv.Atoi(getEnv("TOKEN_CACHE_TTL", "300"))
	enableCache := getEnv("ENABLE_TOKEN_CACHE", "true") == "true"

	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisAddr := redisHost + ":" + redisPort

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
		Redis: RedisConfig{
			Addr:     redisAddr,
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		Performance: PerformanceConfig{
			BcryptCost:    bcryptCost,
			BcryptWorkers: bcryptWorkers,
			TokenCacheTTL: tokenCacheTTL,
			EnableCache:   enableCache,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
