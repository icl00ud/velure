package database

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/icl00ud/velure/services/auth-service/internal/config"
	"github.com/icl00ud/velure/services/auth-service/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var dsn string

	if cfg.URL != "" {
		dsn = cfg.URL
	} else {
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
			cfg.Host, cfg.Username, cfg.Password, cfg.Database, cfg.Port)
	}

	var dialector gorm.Dialector
	if strings.HasPrefix(dsn, "sqlite://") {
		sqliteDSN := strings.TrimPrefix(dsn, "sqlite://")
		dialector = sqlite.Open(sqliteDSN)
	} else {
		dialector = postgres.Open(dsn)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool to avoid exhausting RDS connections.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// SetMaxOpenConns: maximum open connections to the database.
	// Math: 3 pods × 25 conns = 75 total connections (well under the RDS cap).
	sqlDB.SetMaxOpenConns(25)

	// SetMaxIdleConns: maximum idle connections in the pool.
	// Keeps connections ready for quick reuse without wasting resources.
	sqlDB.SetMaxIdleConns(10)

	// SetConnMaxLifetime: maximum time a connection can be reused.
	// Forces periodic recycling to avoid stale connections.
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// SetConnMaxIdleTime: maximum time a connection can stay idle.
	// Closes idle connections more aggressively to free resources.
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	log.Printf("✅ Database connection pool configured: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%s, MaxIdleTime=%s",
		25, 10, 5*time.Minute, 2*time.Minute)

	return db, nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.PasswordReset{},
	)
}
