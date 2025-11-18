package database

import (
	"fmt"
	"time"

	"velure-auth-service/internal/config"
	"velure-auth-service/internal/models"

	"gorm.io/driver/postgres"
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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configurar connection pool para alta performance
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// SetMaxOpenConns: máximo de conexões abertas ao banco
	sqlDB.SetMaxOpenConns(100)

	// SetMaxIdleConns: máximo de conexões idle no pool (aumentado para evitar connection churn)
	sqlDB.SetMaxIdleConns(50)

	// SetConnMaxLifetime: tempo máximo que uma conexão pode ser reusada
	sqlDB.SetConnMaxLifetime(time.Hour)

	// SetConnMaxIdleTime: tempo máximo que uma conexão pode ficar idle
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	return db, nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.PasswordReset{},
	)
}
