package database

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	migratedb "github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

type migrator interface {
	Up() error
	Version() (uint, bool, error)
}

var createDriver = func(db *sql.DB) (migratedb.Driver, error) {
	return postgres.WithInstance(db, &postgres.Config{})
}

var createMigrateInstance = func(driver migratedb.Driver, migrationsPath string) (migrator, error) {
	return migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
}

var newMigrator = func(db *sql.DB, migrationsPath string) (migrator, error) {
	driver, err := createDriver(db)
	if err != nil {
		return nil, fmt.Errorf("create migration driver: %w", err)
	}

	m, err := createMigrateInstance(driver, migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("create migration instance: %w", err)
	}
	return m, nil
}

func RunMigrations(db *sql.DB, migrationsPath string) error {
	m, err := newMigrator(db, migrationsPath)
	if err != nil {
		return err
	}

	version, dirty, _ := m.Version()
	zap.L().Info("current migration state",
		zap.Uint("version", version),
		zap.Bool("dirty", dirty))

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}

	version, dirty, _ = m.Version()
	zap.L().Info("migrations applied successfully",
		zap.Uint("version", version),
		zap.Bool("dirty", dirty))

	return nil
}
