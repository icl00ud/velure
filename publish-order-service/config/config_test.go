package config

import (
	"os"
	"testing"
)

func TestLoad_Success(t *testing.T) {
	os.Setenv("PUBLISH_ORDER_SERVICE_APP_PORT", "8080")
	os.Setenv("POSTGRES_URL", "postgres://u:p@localhost/db?sslmode=disable")
	os.Setenv("RABBITMQ_URL", "amqp://u:p@host:5672/")
	os.Setenv("RABBITMQ_EXCHANGE", "ex")
	defer os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() erro inesperado: %v", err)
	}
	if cfg.Port != "8080" ||
		cfg.PostgresURL != "postgres://u:p@localhost/db?sslmode=disable" ||
		cfg.RabbitURL != "amqp://u:p@host:5672/" ||
		cfg.Exchange != "ex" {
		t.Errorf("cfg incorreto: %+v", cfg)
	}
}

func TestLoad_MissingEnv(t *testing.T) {
	os.Clearenv()
	_, err := Load()
	if err == nil {
		t.Fatal("esperava erro quando falta env, mas retornou nil")
	}
}
