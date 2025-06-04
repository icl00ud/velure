package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port        string
	PostgresURL string
	RabbitURL   string
	Exchange    string
}

func Load() (Config, error) {
	var missing []string
	var c Config

	if v, ok := os.LookupEnv("PUBLISHER_ORDER_SERVICE_APP_PORT"); ok && strings.TrimSpace(v) != "" {
		c.Port = v
	} else {
		missing = append(missing, "PUBLISHER_ORDER_SERVICE_APP_PORT")
	}

        if v, ok := os.LookupEnv("POSTGRES_URL"); ok && strings.TrimSpace(v) != "" {
                c.PostgresURL = v
        } else {
                missing = append(missing, "POSTGRES_URL")
        }

	if v, ok := os.LookupEnv("PUBLISHER_RABBITMQ_URL"); ok && strings.TrimSpace(v) != "" {
		c.RabbitURL = v
	} else {
		missing = append(missing, "PUBLISHER_RABBITMQ_URL")
	}

	if v, ok := os.LookupEnv("ORDER_EXCHANGE"); ok && strings.TrimSpace(v) != "" {
		c.Exchange = v
	} else {
		missing = append(missing, "ORDER_EXCHANGE")
	}

	if len(missing) > 0 {
		return c, fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}
	return c, nil
}
