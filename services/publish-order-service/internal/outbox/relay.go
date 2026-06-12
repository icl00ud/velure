package outbox

import (
	"context"
	"database/sql"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/icl00ud/velure/services/publish-order-service/internal/metrics"
	"github.com/icl00ud/velure/services/publish-order-service/internal/model"
	"github.com/icl00ud/velure/services/publish-order-service/internal/telemetry"
	"github.com/icl00ud/velure/shared/logger"
)

type Publisher interface {
	PublishWithConfirm(ctx context.Context, evt model.OutboxEvent) error
}

type Relay struct {
	repo      Repository
	publisher Publisher
	logger    *logger.Logger
	interval  time.Duration
	batchSize int
	commit    func(*sql.Tx) error
	rollback  func(*sql.Tx) error
	notify    <-chan struct{} // wake signal; nil for polling-only
}

type Option func(*Relay)

func WithInterval(d time.Duration) Option        { return func(r *Relay) { r.interval = d } }
func WithBatchSize(n int) Option                 { return func(r *Relay) { r.batchSize = n } }
func WithLogger(l *logger.Logger) Option         { return func(r *Relay) { r.logger = l } }
func WithNotifyChannel(c <-chan struct{}) Option  { return func(r *Relay) { r.notify = c } }
func WithCommitFn(f func(*sql.Tx) error) Option  { return func(r *Relay) { r.commit = f } }
func WithRollbackFn(f func(*sql.Tx) error) Option { return func(r *Relay) { r.rollback = f } }

func NewRelay(repo Repository, pub Publisher, opts ...Option) *Relay {
	r := &Relay{
		repo:      repo,
		publisher: pub,
		interval:  10 * time.Second,
		batchSize: 50,
		commit:    func(tx *sql.Tx) error { return tx.Commit() },
		rollback:  func(tx *sql.Tx) error { return tx.Rollback() },
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Relay) Start(ctx context.Context) error {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	if err := r.processBatch(ctx); err != nil {
		r.logError("initial batch failed", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := r.processBatch(ctx); err != nil {
				r.logError("poll batch failed", err)
			}
		case <-r.notify:
			if err := r.processBatch(ctx); err != nil {
				r.logError("notify batch failed", err)
			}
		}
	}
}

func (r *Relay) processBatch(ctx context.Context) error {
	start := time.Now()
	// Refresh the gauge even when publishing fails — that is exactly when
	// pending events accumulate and the panel must show it.
	defer func() {
		if n, err := r.repo.CountPending(ctx); err == nil {
			metrics.OutboxEventsPending.Set(float64(n))
		}
	}()
	tx, events, err := r.repo.FetchUnpublished(ctx, r.batchSize)
	if err != nil {
		metrics.OutboxRelayErrors.Inc()
		return err
	}
	if len(events) == 0 {
		_ = r.rollback(tx)
		return nil
	}

	tracer := otel.Tracer("outbox-relay")
	ids := make([]string, 0, len(events))
	for _, evt := range events {
		// Resume the trace of the request that wrote the event, so the AMQP
		// publish shows up as a child of the original HTTP span.
		evtCtx := telemetry.WithTraceparent(ctx, evt.TraceContext)
		evtCtx, span := tracer.Start(evtCtx, "outbox.publish",
			trace.WithSpanKind(trace.SpanKindProducer),
			trace.WithAttributes(
				attribute.String("messaging.destination.name", evt.EventType),
				attribute.String("velure.aggregate_id", evt.AggregateID),
			))
		err := r.publisher.PublishWithConfirm(evtCtx, evt)
		span.End()
		if err != nil {
			_ = r.rollback(tx)
			metrics.OutboxRelayPublished.WithLabelValues("failure").Inc()
			metrics.OutboxRelayErrors.Inc()
			return err
		}
		ids = append(ids, evt.ID)
		metrics.OutboxRelayPublished.WithLabelValues("success").Inc()
	}
	if err := r.repo.MarkPublished(ctx, tx, ids); err != nil {
		_ = r.rollback(tx)
		metrics.OutboxRelayErrors.Inc()
		return err
	}
	if err := r.commit(tx); err != nil {
		metrics.OutboxRelayErrors.Inc()
		return err
	}
	metrics.OutboxRelayBatchDuration.Observe(time.Since(start).Seconds())
	return nil
}

func (r *Relay) logError(msg string, err error) {
	if r.logger == nil {
		return
	}
	r.logger.Error(msg, logger.Err(err))
}
