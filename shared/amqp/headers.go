package amqp

import "github.com/rabbitmq/amqp091-go"

// HeadersToMap converts string-valued AMQP headers (where the W3C trace
// context travels) into the map form the OTel propagator understands.
func HeadersToMap(headers amqp091.Table) map[string]string {
	m := make(map[string]string, len(headers))
	for k, v := range headers {
		if s, ok := v.(string); ok {
			m[k] = s
		}
	}
	return m
}

// RetryCount reads the number of prior delivery attempts from the x-death
// header RabbitMQ attaches when a message is dead-lettered back onto a queue.
func RetryCount(headers amqp091.Table) int64 {
	if headers == nil {
		return 0
	}
	xDeath, ok := headers["x-death"].([]any)
	if !ok || len(xDeath) == 0 {
		return 0
	}
	death, ok := xDeath[0].(amqp091.Table)
	if !ok {
		return 0
	}
	count, ok := death["count"].(int64)
	if !ok {
		return 0
	}
	return count
}
