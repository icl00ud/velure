package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoad_Success(t *testing.T) {
	// Set all required environment variables
	os.Setenv("PROCESS_ORDER_SERVICE_APP_PORT", "8081")
	os.Setenv("PROCESS_RABBITMQ_URL", "amqp://localhost")
	os.Setenv("RABBITMQ_ORDER_QUEUE", "orders")
	os.Setenv("ORDER_EXCHANGE", "order-exchange")
	os.Setenv("WORKERS", "50")
	defer func() {
		os.Unsetenv("PROCESS_ORDER_SERVICE_APP_PORT")
		os.Unsetenv("PROCESS_RABBITMQ_URL")
		os.Unsetenv("RABBITMQ_ORDER_QUEUE")
		os.Unsetenv("ORDER_EXCHANGE")
		os.Unsetenv("WORKERS")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "8081" {
		t.Errorf("expected Port 8081, got %s", cfg.Port)
	}
	if cfg.RabbitURL != "amqp://localhost" {
		t.Errorf("expected RabbitURL amqp://localhost, got %s", cfg.RabbitURL)
	}
	if cfg.OrderQueue != "orders" {
		t.Errorf("expected OrderQueue orders, got %s", cfg.OrderQueue)
	}
	if cfg.OrderExchange != "order-exchange" {
		t.Errorf("expected OrderExchange order-exchange, got %s", cfg.OrderExchange)
	}
	if cfg.Workers != 50 {
		t.Errorf("expected Workers 50, got %d", cfg.Workers)
	}
}

func TestLoad_DefaultWorkers(t *testing.T) {
	// Set all required, no workers specified
	os.Setenv("PROCESS_ORDER_SERVICE_APP_PORT", "8081")
	os.Setenv("PROCESS_RABBITMQ_URL", "amqp://localhost")
	os.Setenv("RABBITMQ_ORDER_QUEUE", "orders")
	os.Setenv("ORDER_EXCHANGE", "order-exchange")
	defer func() {
		os.Unsetenv("PROCESS_ORDER_SERVICE_APP_PORT")
		os.Unsetenv("PROCESS_RABBITMQ_URL")
		os.Unsetenv("RABBITMQ_ORDER_QUEUE")
		os.Unsetenv("ORDER_EXCHANGE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Workers != 100 {
		t.Errorf("expected default Workers 100, got %d", cfg.Workers)
	}
}

func TestLoad_InvalidWorkers(t *testing.T) {
	// Set workers to invalid value
	os.Setenv("PROCESS_ORDER_SERVICE_APP_PORT", "8081")
	os.Setenv("PROCESS_RABBITMQ_URL", "amqp://localhost")
	os.Setenv("RABBITMQ_ORDER_QUEUE", "orders")
	os.Setenv("ORDER_EXCHANGE", "order-exchange")
	os.Setenv("WORKERS", "invalid")
	defer func() {
		os.Unsetenv("PROCESS_ORDER_SERVICE_APP_PORT")
		os.Unsetenv("PROCESS_RABBITMQ_URL")
		os.Unsetenv("RABBITMQ_ORDER_QUEUE")
		os.Unsetenv("ORDER_EXCHANGE")
		os.Unsetenv("WORKERS")
	}()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid workers, got nil")
	}
	if !strings.Contains(err.Error(), "invalid WORKERS value") {
		t.Errorf("expected error to mention invalid WORKERS value, got: %v", err)
	}
}

func TestLoad_MissingPort(t *testing.T) {
	// Missing PORT
	os.Setenv("PROCESS_RABBITMQ_URL", "amqp://localhost")
	os.Setenv("RABBITMQ_ORDER_QUEUE", "orders")
	os.Setenv("ORDER_EXCHANGE", "order-exchange")
	defer func() {
		os.Unsetenv("PROCESS_RABBITMQ_URL")
		os.Unsetenv("RABBITMQ_ORDER_QUEUE")
		os.Unsetenv("ORDER_EXCHANGE")
	}()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing PORT, got nil")
	}
	if !strings.Contains(err.Error(), "PROCESS_ORDER_SERVICE_APP_PORT") {
		t.Errorf("expected error to mention PROCESS_ORDER_SERVICE_APP_PORT, got: %v", err)
	}
}

func TestLoad_MissingRabbitURL(t *testing.T) {
	// Missing RabbitURL
	os.Setenv("PROCESS_ORDER_SERVICE_APP_PORT", "8081")
	os.Setenv("RABBITMQ_ORDER_QUEUE", "orders")
	os.Setenv("ORDER_EXCHANGE", "order-exchange")
	defer func() {
		os.Unsetenv("PROCESS_ORDER_SERVICE_APP_PORT")
		os.Unsetenv("RABBITMQ_ORDER_QUEUE")
		os.Unsetenv("ORDER_EXCHANGE")
	}()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing RabbitURL, got nil")
	}
	if !strings.Contains(err.Error(), "PROCESS_RABBITMQ_URL") {
		t.Errorf("expected error to mention PROCESS_RABBITMQ_URL, got: %v", err)
	}
}

func TestLoad_MissingMultiple(t *testing.T) {
	// Missing multiple required vars
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing env vars, got nil")
	}
	if !strings.Contains(err.Error(), "missing environment variables") {
		t.Errorf("expected error to mention missing environment variables, got: %v", err)
	}
}

func TestLoad_EmptyValues(t *testing.T) {
	// Set to empty strings or whitespace (should be treated as missing)
	os.Setenv("PROCESS_ORDER_SERVICE_APP_PORT", "  ")
	os.Setenv("PROCESS_RABBITMQ_URL", "")
	os.Setenv("RABBITMQ_ORDER_QUEUE", "  ")
	os.Setenv("ORDER_EXCHANGE", "")
	defer func() {
		os.Unsetenv("PROCESS_ORDER_SERVICE_APP_PORT")
		os.Unsetenv("PROCESS_RABBITMQ_URL")
		os.Unsetenv("RABBITMQ_ORDER_QUEUE")
		os.Unsetenv("ORDER_EXCHANGE")
	}()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for empty env vars, got nil")
	}
}

func TestLoad_WorkersWhitespace(t *testing.T) {
	// Workers with only whitespace should default
	os.Setenv("PROCESS_ORDER_SERVICE_APP_PORT", "8081")
	os.Setenv("PROCESS_RABBITMQ_URL", "amqp://localhost")
	os.Setenv("RABBITMQ_ORDER_QUEUE", "orders")
	os.Setenv("ORDER_EXCHANGE", "order-exchange")
	os.Setenv("WORKERS", "   ")
	defer func() {
		os.Unsetenv("PROCESS_ORDER_SERVICE_APP_PORT")
		os.Unsetenv("PROCESS_RABBITMQ_URL")
		os.Unsetenv("RABBITMQ_ORDER_QUEUE")
		os.Unsetenv("ORDER_EXCHANGE")
		os.Unsetenv("WORKERS")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Workers != 100 {
		t.Errorf("expected default Workers 100 for whitespace value, got %d", cfg.Workers)
	}
}
