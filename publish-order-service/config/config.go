package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	PostgresURL string
	RabbitURL   string
	Exchange    string
}

func Load() (Config, error) {
	c := Config{
		Port:        os.Getenv("PUBLISH_ORDER_SERVICE_APP_PORT"),
		PostgresURL: os.Getenv("POSTGRES_URL"),
		RabbitURL:   os.Getenv("RABBITMQ_URL"),
		Exchange:    os.Getenv("RABBITMQ_EXCHANGE"),
	}
	if c.Port == "" || c.PostgresURL == "" || c.RabbitURL == "" || c.Exchange == "" {
		return c, fmt.Errorf("missing required environment variables")
	}
	return c, nil
}
