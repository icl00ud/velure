package outbox

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakePqListener simulates the subset of *pq.Listener we use.
type fakePqListener struct {
	notify chan struct{}
	closed bool
}

func (f *fakePqListener) Listen(channel string) error { return nil }
func (f *fakePqListener) NotificationChannel() <-chan struct{} {
	if f.notify == nil {
		f.notify = make(chan struct{}, 1)
	}
	return f.notify
}
func (f *fakePqListener) Close() error { f.closed = true; return nil }

func TestListener_ForwardsNotifications(t *testing.T) {
	pq := &fakePqListener{notify: make(chan struct{}, 1)}
	l := newListenerWith(pq, "outbox_new")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	out := make(chan struct{}, 1)
	go l.run(ctx, out)

	pq.notify <- struct{}{}

	select {
	case <-out:
		// ok
	case <-time.After(50 * time.Millisecond):
		t.Fatal("expected forwarded notification")
	}
}

func TestListener_StopsOnContextCancel(t *testing.T) {
	pq := &fakePqListener{}
	l := newListenerWith(pq, "outbox_new")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	out := make(chan struct{}, 1)
	err := l.run(ctx, out)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
