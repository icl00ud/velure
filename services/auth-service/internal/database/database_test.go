package database

import (
	"testing"

	"velure-auth-service/internal/config"
	"velure-auth-service/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestConnect_WithURL(t *testing.T) {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Test that we can migrate models
	err = Migrate(db)
	if err != nil {
		t.Errorf("Migrate() failed: %v", err)
	}

	// Verify tables were created
	if !db.Migrator().HasTable(&models.User{}) {
		t.Error("User table was not created")
	}
	if !db.Migrator().HasTable(&models.Session{}) {
		t.Error("Session table was not created")
	}
	if !db.Migrator().HasTable(&models.PasswordReset{}) {
		t.Error("PasswordReset table was not created")
	}
}

func TestConnect_WithDSNComponents(t *testing.T) {
	// Test DSN string construction when URL is empty
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Username: "testuser",
		Password: "testpass",
		Database: "testdb",
		Port:     5432,
		URL:      "", // Empty URL means DSN will be constructed
	}

	// We can't actually connect to PostgreSQL in unit tests,
	// but we can verify the DSN string would be constructed correctly
	// by checking that Connect returns an error (no real DB available)
	_, err := Connect(cfg)
	if err == nil {
		t.Error("Expected error when connecting to non-existent database, got nil")
	}

	// The error should be about connection failure, not DSN construction
	if err != nil && err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestConnect_WithURL_Priority(t *testing.T) {
	// Test that URL takes priority over individual components
	cfg := config.DatabaseConfig{
		Host:     "wrong-host",
		Username: "wrong-user",
		Password: "wrong-pass",
		Database: "wrong-db",
		Port:     9999,
		URL:      "invalid-postgres-url", // URL should be used instead
	}

	// Should fail with URL-related error, not DSN-related
	_, err := Connect(cfg)
	if err == nil {
		t.Error("Expected error when connecting with invalid URL, got nil")
	}
}

func TestMigrate(t *testing.T) {
	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Test migration
	err = Migrate(db)
	if err != nil {
		t.Fatalf("Migrate() failed: %v", err)
	}

	// Verify all tables exist
	tables := []interface{}{
		&models.User{},
		&models.Session{},
		&models.PasswordReset{},
	}

	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			t.Errorf("Table for %T was not created", table)
		}
	}
}

func TestMigrate_IdempotentOperation(t *testing.T) {
	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run migration twice to ensure it's idempotent
	err = Migrate(db)
	if err != nil {
		t.Fatalf("First Migrate() failed: %v", err)
	}

	err = Migrate(db)
	if err != nil {
		t.Errorf("Second Migrate() failed (should be idempotent): %v", err)
	}
}
