package queue

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/icl00ud/process-order-service/internal/client"
	"github.com/icl00ud/process-order-service/internal/model"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const maxRetries = 3

type Consumer interface {
	Consume(ctx context.Context, handler func(model.Event) error) error
	Close() error
}

type rabbitMQConsumer struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	queue   string
	logger  *zap.Logger
}

func NewRabbitMQConsumer(amqpURL, queueName string, logger *zap.Logger) (Consumer, error) {
	conn, err := amqp091.Dial(amqpURL)
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

	return &rabbitMQConsumer{conn: conn, channel: ch, queue: queueName, logger: logger}, nil
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

func (r *rabbitMQConsumer) Consume(ctx context.Context, handler func(model.Event) error) error {
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

			var evt model.Event
			if err := json.Unmarshal(d.Body, &evt); err != nil {
				// Erro de parsing é permanente - envia direto para DLQ
				d.Nack(false, false)
				r.logger.Error("invalid event structure - sending to DLQ",
					zap.Error(err),
					zap.Int64("retry_count", retryCount))
				continue
			}

			r.logger.Info("payment processing started",
				zap.String("event_type", evt.Type),
				zap.Int64("retry_count", retryCount))

			if err := handler(evt); err != nil {
				// Verifica se é erro permanente
				var permErr *client.PermanentError
				if errors.As(err, &permErr) {
					// Erro permanente (ex: produto não encontrado) - envia para DLQ imediatamente
					d.Nack(false, false)
					r.logger.Error("permanent error - sending to DLQ",
						zap.Error(err),
						zap.Int("status_code", permErr.StatusCode),
						zap.Int64("retry_count", retryCount))
					continue
				}

				// Erro temporário - verifica limite de retries
				if retryCount >= maxRetries {
					// Excedeu limite de retries - envia para DLQ
					d.Nack(false, false)
					r.logger.Error("max retries exceeded - sending to DLQ",
						zap.Error(err),
						zap.Int64("retry_count", retryCount),
						zap.Int("max_retries", maxRetries))
					continue
				}

				// Erro temporário com retries disponíveis - requeue
				d.Nack(false, true)
				r.logger.Warn("transient error - requeueing for retry",
					zap.Error(err),
					zap.Int64("retry_count", retryCount),
					zap.Int64("remaining_retries", maxRetries-retryCount))
				continue
			}

			// Sucesso - ACK
			d.Ack(false)
			r.logger.Info("payment processed successfully",
				zap.String("event_type", evt.Type),
				zap.Int64("retry_count", retryCount))
		}
	}
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
