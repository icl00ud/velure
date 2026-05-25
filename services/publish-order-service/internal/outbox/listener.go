package outbox

import (
	"context"
	"time"

	"github.com/icl00ud/velure/services/publish-order-service/internal/metrics"
	"github.com/icl00ud/velure/shared/logger"
	"github.com/lib/pq"
)

// pqListener is the subset of *pq.Listener we depend on; abstracted for tests.
type pqListener interface {
	Listen(channel string) error
	NotificationChannel() <-chan struct{}
	Close() error
}

// Listener subscribes to a Postgres LISTEN/NOTIFY channel and forwards wake
// signals to a caller-provided channel.
type Listener struct {
	pq      pqListener
	channel string
	log     *logger.Logger
}

// NewListener wraps lib/pq's NewListener, subscribing to a Postgres NOTIFY channel.
// The lib/pq listener manages its own connection and reconnect backoff (1s..30s).
func NewListener(dsn, channel string, log *logger.Logger) *Listener {
	pql := pq.NewListener(dsn, time.Second, 30*time.Second, func(ev pq.ListenerEventType, err error) {
		if ev == pq.ListenerEventReconnected {
			metrics.OutboxListenerReconnects.Inc()
		}
		if err != nil && log != nil {
			log.Warn("listener event", logger.Any("event", ev), logger.Err(err))
		}
	})
	return newListenerWith(adaptPqListener{inner: pql, channel: channel}, channel)
}

func newListenerWith(pql pqListener, channel string) *Listener {
	return &Listener{pq: pql, channel: channel}
}

// Start subscribes to the channel and forwards a wake signal on out.
// Returns when ctx is cancelled.
func (l *Listener) Start(ctx context.Context, out chan<- struct{}) error {
	return l.run(ctx, out)
}

func (l *Listener) run(ctx context.Context, out chan<- struct{}) error {
	if err := l.pq.Listen(l.channel); err != nil {
		return err
	}
	defer l.pq.Close()

	in := l.pq.NotificationChannel()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-in:
			// Coalesce: if there's already a pending wake, drop the duplicate.
			select {
			case out <- struct{}{}:
			default:
			}
		}
	}
}

// adaptPqListener bridges *pq.Listener to our pqListener interface.
// pq.Listener.Notify is a chan *pq.Notification; we coerce to chan struct{}.
//
// NOTE: the goroutine started by NotificationChannel will not exit until
// a.inner.Notify is closed. lib/pq does NOT close Notify on Close(), so
// this goroutine may leak after Close(). Acceptable for the current use-case
// (one long-lived listener per process); tracked as future work.
type adaptPqListener struct {
	inner   *pq.Listener
	channel string
}

func (a adaptPqListener) Listen(ch string) error { return a.inner.Listen(ch) }

func (a adaptPqListener) NotificationChannel() <-chan struct{} {
	out := make(chan struct{}, 16)
	go func() {
		for n := range a.inner.Notify {
			_ = n
			select {
			case out <- struct{}{}:
			default:
			}
		}
		close(out)
	}()
	return out
}

func (a adaptPqListener) Close() error { return a.inner.Close() }
