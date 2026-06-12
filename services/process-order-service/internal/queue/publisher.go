package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/icl00ud/velure/services/process-order-service/internal/model"
	"github.com/icl00ud/velure/services/process-order-service/internal/telemetry"
	"github.com/icl00ud/velure/shared/logger"
	"github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	Publish(ctx context.Context, evt model.Event) error
	Close() error
}

type rabbitPublisher struct {
	conn     AMQPConnection
	channel  AMQPChannel
	exchange string
	logger   *logger.Logger

	// reconnect re-establishes a channel after the broker drops the
	// connection; nil disables the retry (tests without a broker).
	reconnect func() (AMQPChannel, error)
}

func NewRabbitPublisher(amqpURL, exchange string, log *logger.Logger) (Publisher, error) {
	conn, err := amqpDial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	return &rabbitPublisher{conn: conn, channel: ch, exchange: exchange, logger: log}, nil
}

func (r *rabbitPublisher) Publish(ctx context.Context, evt model.Event) error {
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	// Propagate the trace context to downstream consumers via AMQP headers.
	headers := amqp091.Table{}
	for k, v := range telemetry.InjectMap(ctx) {
		headers[k] = v
	}
	publishFunc := func() error {
		return r.channel.Publish(
			r.exchange,
			evt.Type,
			false, false,
			amqp091.Publishing{ContentType: "application/json", Body: body, Headers: headers},
		)
	}

	// A dead channel after a broker restart is permanent for this process;
	// without a retry, status events (order.failed/completed) are lost and
	// orders stay CREATED forever.
	if err := publishFunc(); err != nil {
		if r.reconnect == nil {
			r.logger.Error("publish failed", logger.Err(err), logger.String("exchange", r.exchange), logger.String("event_type", evt.Type))
			return err
		}
		r.logger.Warn("publish failed, attempting reconnect", logger.Err(err), logger.String("event_type", evt.Type))
		ch, recErr := r.reconnect()
		if recErr != nil {
			r.logger.Error("reconnect failed", logger.Err(recErr))
			return err
		}
		r.channel = ch
		if err := publishFunc(); err != nil {
			r.logger.Error("publish failed after reconnect", logger.Err(err), logger.String("event_type", evt.Type))
			return err
		}
	}
	r.logger.Info("payment event published", logger.String("exchange", r.exchange), logger.String("event_type", evt.Type))
	return nil
}

func (r *rabbitPublisher) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
