package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoad_Success(t *testing.T) {
	// Set all required environment variables
	os.Setenv("PUBLISHER_ORDER_SERVICE_APP_PORT", "8080")
	os.Setenv("POSTGRES_URL", "postgres://localhost/testdb")
	os.Setenv("PUBLISHER_RABBITMQ_URL", "amqp://localhost")
	os.Setenv("ORDER_EXCHANGE", "orders")
	os.Setenv("PUBLISHER_ORDER_QUEUE", "test-queue")
	os.Setenv("PUBLISHER_CONSUMER_WORKERS", "5")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("PUBLISHER_ORDER_SERVICE_APP_PORT")
		os.Unsetenv("POSTGRES_URL")
		os.Unsetenv("PUBLISHER_RABBITMQ_URL")
		os.Unsetenv("ORDER_EXCHANGE")
		os.Unsetenv("PUBLISHER_ORDER_QUEUE")
		os.Unsetenv("PUBLISHER_CONSUMER_WORKERS")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("expected Port 8080, got %s", cfg.Port)
	}
	if cfg.PostgresURL != "postgres://localhost/testdb" {
		t.Errorf("expected PostgresURL postgres://localhost/testdb, got %s", cfg.PostgresURL)
	}
	if cfg.RabbitURL != "amqp://localhost" {
		t.Errorf("expected RabbitURL amqp://localhost, got %s", cfg.RabbitURL)
	}
	if cfg.Exchange != "orders" {
		t.Errorf("expected Exchange orders, got %s", cfg.Exchange)
	}
	if cfg.Queue != "test-queue" {
		t.Errorf("expected Queue test-queue, got %s", cfg.Queue)
	}
	if cfg.Workers != 5 {
		t.Errorf("expected Workers 5, got %d", cfg.Workers)
	}
	if cfg.JWTSecret != "test-secret" {
		t.Errorf("expected JWTSecret test-secret, got %s", cfg.JWTSecret)
	}
}

func TestLoad_DefaultQueue(t *testing.T) {
	// Set all required except PUBLISHER_ORDER_QUEUE
	os.Setenv("PUBLISHER_ORDER_SERVICE_APP_PORT", "8080")
	os.Setenv("POSTGRES_URL", "postgres://localhost/testdb")
	os.Setenv("PUBLISHER_RABBITMQ_URL", "amqp://localhost")
	os.Setenv("ORDER_EXCHANGE", "orders")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("PUBLISHER_ORDER_SERVICE_APP_PORT")
		os.Unsetenv("POSTGRES_URL")
		os.Unsetenv("PUBLISHER_RABBITMQ_URL")
		os.Unsetenv("ORDER_EXCHANGE")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Queue != "publish-order-status-updates" {
		t.Errorf("expected default Queue publish-order-status-updates, got %s", cfg.Queue)
	}
}

func TestLoad_DefaultWorkers(t *testing.T) {
	// Set all required, no workers specified
	os.Setenv("PUBLISHER_ORDER_SERVICE_APP_PORT", "8080")
	os.Setenv("POSTGRES_URL", "postgres://localhost/testdb")
	os.Setenv("PUBLISHER_RABBITMQ_URL", "amqp://localhost")
	os.Setenv("ORDER_EXCHANGE", "orders")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("PUBLISHER_ORDER_SERVICE_APP_PORT")
		os.Unsetenv("POSTGRES_URL")
		os.Unsetenv("PUBLISHER_RABBITMQ_URL")
		os.Unsetenv("ORDER_EXCHANGE")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Workers != 3 {
		t.Errorf("expected default Workers 3, got %d", cfg.Workers)
	}
}

func TestLoad_InvalidWorkers(t *testing.T) {
	// Set workers to invalid value
	os.Setenv("PUBLISHER_ORDER_SERVICE_APP_PORT", "8080")
	os.Setenv("POSTGRES_URL", "postgres://localhost/testdb")
	os.Setenv("PUBLISHER_RABBITMQ_URL", "amqp://localhost")
	os.Setenv("ORDER_EXCHANGE", "orders")
	os.Setenv("PUBLISHER_CONSUMER_WORKERS", "invalid")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("PUBLISHER_ORDER_SERVICE_APP_PORT")
		os.Unsetenv("POSTGRES_URL")
		os.Unsetenv("PUBLISHER_RABBITMQ_URL")
		os.Unsetenv("ORDER_EXCHANGE")
		os.Unsetenv("PUBLISHER_CONSUMER_WORKERS")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should default to 3 when invalid
	if cfg.Workers != 3 {
		t.Errorf("expected default Workers 3 for invalid value, got %d", cfg.Workers)
	}
}

func TestLoad_ZeroWorkers(t *testing.T) {
	// Set workers to 0
	os.Setenv("PUBLISHER_ORDER_SERVICE_APP_PORT", "8080")
	os.Setenv("POSTGRES_URL", "postgres://localhost/testdb")
	os.Setenv("PUBLISHER_RABBITMQ_URL", "amqp://localhost")
	os.Setenv("ORDER_EXCHANGE", "orders")
	os.Setenv("PUBLISHER_CONSUMER_WORKERS", "0")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("PUBLISHER_ORDER_SERVICE_APP_PORT")
		os.Unsetenv("POSTGRES_URL")
		os.Unsetenv("PUBLISHER_RABBITMQ_URL")
		os.Unsetenv("ORDER_EXCHANGE")
		os.Unsetenv("PUBLISHER_CONSUMER_WORKERS")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should default to 3 when 0
	if cfg.Workers != 3 {
		t.Errorf("expected default Workers 3 for zero value, got %d", cfg.Workers)
	}
}

func TestLoad_MissingPort(t *testing.T) {
	// Missing PORT
	os.Setenv("POSTGRES_URL", "postgres://localhost/testdb")
	os.Setenv("PUBLISHER_RABBITMQ_URL", "amqp://localhost")
	os.Setenv("ORDER_EXCHANGE", "orders")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("POSTGRES_URL")
		os.Unsetenv("PUBLISHER_RABBITMQ_URL")
		os.Unsetenv("ORDER_EXCHANGE")
		os.Unsetenv("JWT_SECRET")
	}()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing PORT, got nil")
	}
	if !strings.Contains(err.Error(), "PUBLISHER_ORDER_SERVICE_APP_PORT") {
		t.Errorf("expected error to mention PUBLISHER_ORDER_SERVICE_APP_PORT, got: %v", err)
	}
}

func TestLoad_MissingMultiple(t *testing.T) {
	// Missing multiple required vars
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing env vars, got nil")
	}
	if !strings.Contains(err.Error(), "missing required env vars") {
		t.Errorf("expected error to mention missing required env vars, got: %v", err)
	}
}

func TestLoad_EmptyValues(t *testing.T) {
	// Set to empty strings (should be treated as missing)
	os.Setenv("PUBLISHER_ORDER_SERVICE_APP_PORT", "  ")
	os.Setenv("POSTGRES_URL", "")
	os.Setenv("PUBLISHER_RABBITMQ_URL", "  ")
	os.Setenv("ORDER_EXCHANGE", "")
	os.Setenv("JWT_SECRET", "  ")
	defer func() {
		os.Unsetenv("PUBLISHER_ORDER_SERVICE_APP_PORT")
		os.Unsetenv("POSTGRES_URL")
		os.Unsetenv("PUBLISHER_RABBITMQ_URL")
		os.Unsetenv("ORDER_EXCHANGE")
		os.Unsetenv("JWT_SECRET")
	}()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for empty env vars, got nil")
	}
}
