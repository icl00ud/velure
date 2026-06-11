package queue

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/icl00ud/velure/services/process-order-service/internal/client"
	"github.com/icl00ud/velure/services/process-order-service/internal/metrics"
	"github.com/icl00ud/velure/services/process-order-service/internal/model"
	"github.com/icl00ud/velure/services/process-order-service/internal/telemetry"
	"github.com/icl00ud/velure/shared/logger"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// headersToMap converts string-valued AMQP headers (where the W3C trace
// context travels) into the map form the OTel propagator understands.
func headersToMap(headers amqp091.Table) map[string]string {
	m := make(map[string]string, len(headers))
	for k, v := range headers {
		if s, ok := v.(string); ok {
			m[k] = s
		}
	}
	return m
}

const maxRetries = 3

type Consumer interface {
	Consume(ctx context.Context, handler func(ctx context.Context, eventID string, evt model.Event) error) error
	Close() error
}

type amqpAcker interface {
	Ack(tag uint64, multiple bool) error
	Nack(tag uint64, multiple bool, requeue bool) error
}

type rabbitMQConsumer struct {
	conn    AMQPConnection
	channel AMQPChannel
	queue   string
	logger  *logger.Logger
}

func NewRabbitMQConsumer(amqpURL, queueName string, log *logger.Logger) (Consumer, error) {
	conn, err := amqpDial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Não redeclarar a fila aqui - ela é criada pelo bootstrap.sh do RabbitMQ
	// com argumentos específicos (DLX, etc). Redeclarar causaria PRECONDITION_FAILED.
	// _, err = ch.QueueDeclare(queueName, true, false, false, false, nil)

	if err := ch.Qos(50, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &rabbitMQConsumer{conn: conn, channel: ch, queue: queueName, logger: log}, nil
}

// getRetryCount obtém o número de tentativas do header x-death do RabbitMQ
func getRetryCount(headers amqp091.Table) int64 {
	if headers == nil {
		return 0
	}

	xDeath, ok := headers["x-death"].([]interface{})
	if !ok || len(xDeath) == 0 {
		return 0
	}

	// Pega a primeira entrada de x-death (mais recente)
	death, ok := xDeath[0].(amqp091.Table)
	if !ok {
		return 0
	}

	// Conta quantas vezes foi rejeitada
	count, ok := death["count"].(int64)
	if !ok {
		return 0
	}

	return count
}

func (r *rabbitMQConsumer) Consume(ctx context.Context, handler func(ctx context.Context, eventID string, evt model.Event) error) error {
	msgs, err := r.channel.Consume(r.queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-msgs:
			if !ok {
				return nil
			}

			// Contagem de retries da mensagem
			retryCount := getRetryCount(d.Headers)
			eventID := extractEventID(d)

			var evt model.Event
			if err := json.Unmarshal(d.Body, &evt); err != nil {
				// Erro de parsing é permanente - envia direto para DLQ
				d.Nack(false, false)
				r.logger.Error("invalid event structure - sending to DLQ",
					logger.Err(err),
					logger.Int64("retry_count", retryCount))
				continue
			}

			r.logger.Info("payment processing started",
				logger.String("event_type", evt.Type),
				logger.Int64("retry_count", retryCount))

			// Continue the trace propagated through AMQP headers.
			msgCtx := telemetry.ExtractMap(ctx, headersToMap(d.Headers))
			msgCtx, span := otel.Tracer("order-consumer").Start(msgCtx, "consume "+evt.Type,
				trace.WithSpanKind(trace.SpanKindConsumer))

			err := handler(msgCtx, eventID, evt)
			span.End()
			if err != nil {
				// Check whether this is a permanent error
				var permErr *client.PermanentError
				if errors.As(err, &permErr) {
					// Permanent error (e.g. product not found) — route directly to the DLQ
					d.Nack(false, false)
					r.logger.Error("permanent error - sending to DLQ",
						logger.Err(err),
						logger.Int("status_code", permErr.StatusCode),
						logger.Int64("retry_count", retryCount))
					continue
				}

				// Erro temporário - verifica limite de retries
				if retryCount >= maxRetries {
					// Excedeu limite de retries - envia para DLQ
					d.Nack(false, false)
					r.logger.Error("max retries exceeded - sending to DLQ",
						logger.Err(err),
						logger.Int64("retry_count", retryCount),
						logger.Int("max_retries", maxRetries))
					continue
				}

				// Erro temporário com retries disponíveis - requeue
				d.Nack(false, true)
				r.logger.Warn("transient error - requeueing for retry",
					logger.Err(err),
					logger.Int64("retry_count", retryCount),
					logger.Int64("remaining_retries", int64(maxRetries)-retryCount))
				continue
			}

			// Sucesso - ACK
			d.Ack(false)
			r.logger.Info("payment processed",
				logger.String("event_type", evt.Type),
				logger.Int64("retry_count", retryCount))
		}
	}
}

func extractEventID(d amqp091.Delivery) string {
	if d.Headers != nil {
		if v, ok := d.Headers["event_id"].(string); ok && v != "" {
			return v
		}
	}
	if d.MessageId != "" {
		return d.MessageId
	}
	// Neither header nor envelope id present — flag and fall back to a
	// deterministic hash of the body so dedup still works.
	metrics.MessagesMissingEventID.Inc()
	h := sha256.Sum256(d.Body)
	return hex.EncodeToString(h[:8])
}

func (r *rabbitMQConsumer) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
