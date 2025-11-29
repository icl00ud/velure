package database

import (
	"fmt"
	"log"
	"strings"
	"time"

	"velure-auth-service/internal/config"
	"velure-auth-service/internal/models"

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

	// Configurar connection pool para evitar esgotamento de conexões RDS
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// SetMaxOpenConns: máximo de conexões abertas ao banco
	// Cálculo: 3 pods × 25 conns = 75 conexões totais (bem abaixo do limite do RDS)
	sqlDB.SetMaxOpenConns(25)

	// SetMaxIdleConns: máximo de conexões idle no pool
	// Mantém conexões prontas para uso rápido sem desperdiçar recursos
	sqlDB.SetMaxIdleConns(10)

	// SetConnMaxLifetime: tempo máximo que uma conexão pode ser reusada
	// Força reciclagem periódica de conexões para evitar conexões stale
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// SetConnMaxIdleTime: tempo máximo que uma conexão pode ficar idle
	// Fecha conexões ociosas mais rapidamente para liberar recursos
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
