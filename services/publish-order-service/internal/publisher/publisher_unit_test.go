package publisher

import (
	"errors"
	"sync"
	"testing"

	"github.com/icl00ud/publish-order-service/internal/model"
	"github.com/rabbitmq/amqp091-go"
	"github.com/icl00ud/velure-shared/logger"
)

func TestPublish_ReturnsErrorWhenClosed(t *testing.T) {
	pub := &rabbitMQPublisher{
		logger: logger.NewNop(),
		closed: true,
	}

	err := pub.Publish(model.Event{Type: "order.test", Payload: []byte(`{}`)})
	if err == nil {
		t.Fatal("expected error when publishing after Close")
	}
}

func TestClose_IsIdempotentWithoutConnections(t *testing.T) {
	pub := &rabbitMQPublisher{
		logger: logger.NewNop(),
	}

	if err := pub.Close(); err != nil {
		t.Fatalf("first close returned error: %v", err)
	}
	if err := pub.Close(); err != nil {
		t.Fatalf("second close returned error: %v", err)
	}
}

func TestDialPublisher_UsesDialer(t *testing.T) {
	origDial := amqpDialer
	defer func() { amqpDialer = origDial }()

	called := 0
	amqpDialer = func(url string) (*amqp091.Connection, error) {
		called++
		return nil, errors.New("dial fail")
	}

	if _, err := dialPublisher("amqp://example"); err == nil {
		t.Fatal("expected error from dialPublisher")
	}
	if called != 1 {
		t.Fatalf("expected dialer to be invoked once, got %d", called)
	}
}

type fakePubChannel struct {
	publishErr error
	published  int
	closed     bool
	declared   int
	name       string
	closeErr   error
	declareErr error
}

func (f *fakePubChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
	f.published++
	return f.publishErr
}
func (f *fakePubChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp091.Table) error {
	f.declared++
	f.name = name
	return f.declareErr
}
func (f *fakePubChannel) Close() error {
	f.closed = true
	return f.closeErr
}

type fakePubConn struct {
	ch       *fakePubChannel
	closed   bool
	closeErr error
}

func (f *fakePubConn) Channel() (amqpPublisherChannel, error) { return f.ch, nil }
func (f *fakePubConn) Close() error {
	f.closed = true
	return f.closeErr
}

type stubRawConnection struct {
	channelErr error
	closed     bool
}

func (s *stubRawConnection) Channel() (*amqp091.Channel, error) { return nil, s.channelErr }
func (s *stubRawConnection) Close() error {
	s.closed = true
	return nil
}

type failingConn struct {
	closed bool
}

func (f *failingConn) Channel() (amqpPublisherChannel, error) {
	return nil, errors.New("channel failed")
}
func (f *failingConn) Close() error {
	f.closed = true
	return nil
}

func TestPublish_ReconnectsOnError(t *testing.T) {
	first := &fakePubChannel{publishErr: errors.New("fail")}
	second := &fakePubChannel{}
	pub := &rabbitMQPublisher{}
	pub.logger = logger.NewNop()
	pub.exchange = "ex"
	pub.ch = first
	pub.connectFn = func() error {
		pub.ch = second
		return nil
	}

	err := pub.Publish(model.Event{Type: "order.created", Payload: []byte(`{}`)})
	if err != nil {
		t.Fatalf("expected publish to succeed after reconnect, got %v", err)
	}
	if first.published != 1 {
		t.Fatalf("expected first channel to be used once, got %d", first.published)
	}
	if second.published != 1 {
		t.Fatalf("expected second channel to publish once, got %d", second.published)
	}
}

func TestPublish_ReconnectFailure(t *testing.T) {
	ch := &fakePubChannel{publishErr: errors.New("fail")}
	pub := &rabbitMQPublisher{
		logger:   logger.NewNop(),
		exchange: "ex",
		ch:       ch,
		connectFn: func() error {
			return errors.New("reconnect fail")
		},
	}

	err := pub.Publish(model.Event{Type: "order.created", Payload: []byte(`{}`)})
	if err == nil {
		t.Fatal("expected error when reconnect fails")
	}
}

func TestPublish_FailsAfterReconnect(t *testing.T) {
	first := &fakePubChannel{publishErr: errors.New("first")}
	second := &fakePubChannel{publishErr: errors.New("second")}

	pub := &rabbitMQPublisher{
		logger:   logger.NewNop(),
		exchange: "ex",
		ch:       first,
	}
	pub.connectFn = func() error {
		pub.ch = second
		return nil
	}

	err := pub.Publish(model.Event{Type: "order.created", Payload: []byte(`{}`)})
	if err == nil {
		t.Fatal("expected error when publish fails after reconnect")
	}
	if second.published != 1 {
		t.Fatalf("expected second channel to attempt publish, got %d", second.published)
	}
}

func TestPublish_ReconnectsWhenChannelNil(t *testing.T) {
	// Build publisher with nil channel to force reconnect path
	pub := &rabbitMQPublisher{
		logger:  logger.NewNop(),
		amqpURL: "amqp://invalid", // will make connect() fail
		mu:      sync.Mutex{},
		connectFn: func() error {
			return errors.New("dial fail")
		},
	}

	pub.ch = nil
	pub.conn = nil

	err := pub.Publish(model.Event{Type: "order.test", Payload: []byte(`{}`)})
	if err == nil {
		t.Fatal("expected error when reconnect fails")
	}
}

func TestNewRabbitMQPublisher_UsesInjectedDial(t *testing.T) {
	ch := &fakePubChannel{}
	conn := &fakePubConn{ch: ch}
	dialCalled := 0

	pub, err := newRabbitMQPublisher("amqp://example", "ex.test", logger.NewNop(), func(url string) (amqpPublisherConn, error) {
		dialCalled++
		return conn, nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	rmq := pub.(*rabbitMQPublisher)
	if dialCalled != 1 {
		t.Fatalf("expected dial to be called once, got %d", dialCalled)
	}
	if rmq.ch != ch || rmq.conn != conn {
		t.Fatal("publisher did not store connection and channel")
	}
	if ch.declared != 1 || ch.name != "ex.test" {
		t.Fatalf("expected exchange declared once with name ex.test, got %d (%s)", ch.declared, ch.name)
	}
}

func TestRabbitMQPublisher_ConnectClosesOldResources(t *testing.T) {
	oldCh := &fakePubChannel{}
	oldConn := &fakePubConn{}
	newCh := &fakePubChannel{}
	newConn := &fakePubConn{ch: newCh}

	pub := &rabbitMQPublisher{
		logger:   logger.NewNop(),
		amqpURL:  "amqp://example",
		exchange: "ex.retry",
		ch:       oldCh,
		conn:     oldConn,
		dialFn: func(url string) (amqpPublisherConn, error) {
			return newConn, nil
		},
	}

	if err := pub.connect(); err != nil {
		t.Fatalf("connect returned error: %v", err)
	}
	if !oldCh.closed || !oldConn.closed {
		t.Fatal("expected previous channel and connection to be closed")
	}
	if pub.ch != newCh {
		t.Fatal("expected channel to be replaced with new one")
	}
}

func TestNewRabbitMQPublisher_PropagatesConnectError(t *testing.T) {
	if _, err := newRabbitMQPublisher("amqp://example", "ex", logger.NewNop(), func(string) (amqpPublisherConn, error) {
		return nil, errors.New("dial failed")
	}); err == nil {
		t.Fatal("expected error when connect fails")
	}
}

func TestRabbitMQPublisher_ConnectDialError(t *testing.T) {
	pub := &rabbitMQPublisher{
		logger:   logger.NewNop(),
		amqpURL:  "amqp://example",
		exchange: "ex",
		dialFn: func(string) (amqpPublisherConn, error) {
			return nil, errors.New("dial failed")
		},
	}
	if err := pub.connect(); err == nil {
		t.Fatal("expected error when dial fails")
	}
}

func TestRabbitMQPublisher_ChannelOpenError(t *testing.T) {
	conn := &failingConn{}
	pub := &rabbitMQPublisher{
		logger:   logger.NewNop(),
		amqpURL:  "amqp://example",
		exchange: "ex",
		dialFn: func(string) (amqpPublisherConn, error) {
			return conn, nil
		},
	}
	if err := pub.connect(); err == nil {
		t.Fatal("expected error when channel open fails")
	}
	if !conn.closed {
		t.Fatal("expected connection to close on channel failure")
	}
}

func TestRabbitMQPublisher_ExchangeDeclareError(t *testing.T) {
	ch := &fakePubChannel{declareErr: errors.New("exchange declare fail")}
	conn := &fakePubConn{ch: ch}
	pub := &rabbitMQPublisher{
		logger:   logger.NewNop(),
		amqpURL:  "amqp://example",
		exchange: "ex",
		dialFn: func(string) (amqpPublisherConn, error) {
			return conn, nil
		},
	}

	if err := pub.connect(); err == nil {
		t.Fatal("expected error when exchange declaration fails")
	}
	if !ch.closed || !conn.closed {
		t.Fatal("expected channel and connection closed on declare failure")
	}
}

func TestNewRabbitMQPublisher_UsesDialVariable(t *testing.T) {
	origDial := dialPublisher
	defer func() { dialPublisher = origDial }()

	ch := &fakePubChannel{}
	conn := &fakePubConn{ch: ch}
	dialPublisher = func(amqpURL string) (amqpPublisherConn, error) {
		return conn, nil
	}

	pub, err := NewRabbitMQPublisher("amqp://example", "ex.var", logger.NewNop())
	if err != nil {
		t.Fatalf("expected constructor to succeed, got %v", err)
	}
	if pub.(*rabbitMQPublisher).ch != ch {
		t.Fatal("expected publisher to keep created channel")
	}
}

func TestLivePublisherConn_AllowsNil(t *testing.T) {
	conn := &livePublisherConn{}
	if _, err := conn.Channel(); err == nil {
		t.Fatal("expected error when channel requested on nil connection")
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("expected nil error closing nil connection, got %v", err)
	}
}

func TestLivePublisherConn_Delegates(t *testing.T) {
	raw := &stubRawConnection{channelErr: errors.New("boom")}
	conn := &livePublisherConn{conn: raw}
	if _, err := conn.Channel(); err == nil {
		t.Fatal("expected error from underlying channel")
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("expected nil error closing connection, got %v", err)
	}
	if !raw.closed {
		t.Fatal("expected underlying close to be called")
	}
}

func TestClose_ClosesChannelAndConnection(t *testing.T) {
	ch := &fakePubChannel{}
	conn := &fakePubConn{}
	pub := &rabbitMQPublisher{
		logger: logger.NewNop(),
		ch:     ch,
		conn:   conn,
	}

	if err := pub.Close(); err != nil {
		t.Fatalf("unexpected error closing publisher: %v", err)
	}
	if !ch.closed || !conn.closed {
		t.Fatal("expected channel and connection to be closed")
	}
}

func TestClose_ReturnsConnectionError(t *testing.T) {
	ch := &fakePubChannel{closeErr: errors.New("channel close")}
	conn := &fakePubConn{closeErr: errors.New("conn close")}
	pub := &rabbitMQPublisher{
		logger: logger.NewNop(),
		ch:     ch,
		conn:   conn,
	}

	if err := pub.Close(); err == nil {
		t.Fatal("expected error when connection close fails")
	}
	if !ch.closed || !conn.closed {
		t.Fatal("expected resources to be closed even on error")
	}
}
