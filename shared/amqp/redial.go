// Package amqp holds RabbitMQ helpers shared by services that consume or
// publish AMQP events, so reconnect/backoff logic lives in one place instead
// of being reimplemented per service.
package amqp

import (
	"context"
	"time"
)

// DefaultReconnectDelay is used by Redial when the caller passes delay <= 0.
const DefaultReconnectDelay = 5 * time.Second

// Redial retries connect until it succeeds or ctx is cancelled, waiting delay
// between attempts (DefaultReconnectDelay if delay <= 0). onError is called
// with each failed attempt's error before the wait.
func Redial[T any](ctx context.Context, delay time.Duration, onError func(error), connect func(ctx context.Context) (T, error)) (T, error) {
	if delay <= 0 {
		delay = DefaultReconnectDelay
	}
	for {
		v, err := connect(ctx)
		if err == nil {
			return v, nil
		}
		onError(err)
		select {
		case <-ctx.Done():
			var zero T
			return zero, ctx.Err()
		case <-time.After(delay):
		}
	}
}
