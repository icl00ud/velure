package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port          string
	RabbitURL     string
	OrderQueue    string
	OrderExchange string
	Workers       int
}

func Load() (Config, error) {
	var missing []string
	var c Config

	if v := strings.TrimSpace(os.Getenv("PROCESS_ORDER_SERVICE_APP_PORT")); v != "" {
		c.Port = v
	} else {
		missing = append(missing, "PROCESS_ORDER_SERVICE_APP_PORT")
	}

	if v := strings.TrimSpace(os.Getenv("PROCESS_RABBITMQ_URL")); v != "" {
		c.RabbitURL = v
	} else {
		missing = append(missing, "PROCESS_RABBITMQ_URL")
	}

	if v := strings.TrimSpace(os.Getenv("RABBITMQ_ORDER_QUEUE")); v != "" {
		c.OrderQueue = v
	} else {
		missing = append(missing, "RABBITMQ_ORDER_QUEUE")
	}

	if v := strings.TrimSpace(os.Getenv("ORDER_EXCHANGE")); v != "" {
		c.OrderExchange = v
	} else {
		missing = append(missing, "ORDER_EXCHANGE")
	}

	if v := strings.TrimSpace(os.Getenv("WORKERS")); v != "" {
		w, err := strconv.Atoi(v)
		if err != nil {
			return c, fmt.Errorf("invalid WORKERS value: %w", err)
		}
		c.Workers = w
	} else {
		c.Workers = 10
	}

	if len(missing) > 0 {
		return c, fmt.Errorf("missing environment variables: %s", strings.Join(missing, ", "))
	}
	return c, nil
}
