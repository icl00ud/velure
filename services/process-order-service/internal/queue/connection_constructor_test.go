package queue

import (
	"errors"
	"testing"

	"github.com/icl00ud/velure-shared/logger"
	"github.com/rabbitmq/amqp091-go"
)

type stubAMQPChannel struct {
	consumeCh     <-chan amqp091.Delivery
	consumeErr    error
	qosErr        error
	bindErr       error
	declareErr    error
	publishErr    error
	qosCalled     bool
	bindCalled    bool
	declareCalled bool
	closed        bool
}

func (s *stubAMQPChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error) {
	if s.consumeErr != nil {
		return nil, s.consumeErr
	}
	if s.consumeCh == nil {
		ch := make(chan amqp091.Delivery)
		close(ch)
		s.consumeCh = ch
	}
	return s.consumeCh, nil
}

func (s *stubAMQPChannel) Qos(prefetchCount, prefetchSize int, global bool) error {
	s.qosCalled = true
	return s.qosErr
}

func (s *stubAMQPChannel) QueueBind(queue, key, exchange string, noWait bool, args amqp091.Table) error {
	s.bindCalled = true
	return s.bindErr
}

func (s *stubAMQPChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
	return s.publishErr
}

func (s *stubAMQPChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp091.Table) error {
	s.declareCalled = true
	return s.declareErr
}

func (s *stubAMQPChannel) Close() error {
	s.closed = true
	return nil
}

type stubAMQPConnection struct {
	ch         AMQPChannel
	channelErr error
	closed     bool
}

func (s *stubAMQPConnection) Channel() (AMQPChannel, error) {
	if s.channelErr != nil {
		return nil, s.channelErr
	}
	return s.ch, nil
}

func (s *stubAMQPConnection) Close() error {
	s.closed = true
	return nil
}

func TestNewRabbitMQConnectionUsesDialer(t *testing.T) {
	stubConn := &stubAMQPConnection{}
	restore := SetAMQPDialer(func(url string) (AMQPConnection, error) {
		if url != "amqp://stub" {
			t.Fatalf("unexpected url %s", url)
		}
		return stubConn, nil
	})
	defer restore()

	conn, err := NewRabbitMQConnection("amqp://stub", logger.NewNop())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if conn.conn != stubConn {
		t.Fatalf("expected stub connection to be stored")
	}
}

func TestRabbitMQConnection_NewConsumer_SetsQosAndBinding(t *testing.T) {
	channel := &stubAMQPChannel{}
	stubConn := &stubAMQPConnection{ch: channel}
	rc := &RabbitMQConnection{conn: stubConn, logger: logger.NewNop()}

	consumer, err := rc.NewConsumer("orders")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if consumer == nil {
		t.Fatal("expected consumer instance")
	}
	if !channel.bindCalled {
		t.Fatal("expected queue bind to be called")
	}
	if !channel.qosCalled {
		t.Fatal("expected qos to be called")
	}
}

func TestRabbitMQConnection_NewConsumer_BindErrorClosesChannel(t *testing.T) {
	channel := &stubAMQPChannel{bindErr: errors.New("bind error")}
	stubConn := &stubAMQPConnection{ch: channel}
	rc := &RabbitMQConnection{conn: stubConn, logger: logger.NewNop()}

	if _, err := rc.NewConsumer("orders"); err == nil {
		t.Fatal("expected bind error")
	}
	if !channel.closed {
		t.Fatal("expected channel to be closed on bind error")
	}
}

func TestRabbitMQConnection_NewConsumer_QosErrorClosesChannel(t *testing.T) {
	channel := &stubAMQPChannel{qosErr: errors.New("qos error")}
	stubConn := &stubAMQPConnection{ch: channel}
	rc := &RabbitMQConnection{conn: stubConn, logger: logger.NewNop()}

	if _, err := rc.NewConsumer("orders"); err == nil {
		t.Fatal("expected qos error")
	}
	if !channel.closed {
		t.Fatal("expected channel to be closed on qos error")
	}
}

func TestRabbitMQConnection_NewPublisher_DeclareErrorClosesChannel(t *testing.T) {
	channel := &stubAMQPChannel{declareErr: errors.New("declare error")}
	stubConn := &stubAMQPConnection{ch: channel}
	rc := &RabbitMQConnection{conn: stubConn, logger: logger.NewNop()}

	if _, err := rc.NewPublisher("orders"); err == nil {
		t.Fatal("expected declare error")
	}
	if !channel.closed {
		t.Fatal("expected channel to be closed on declare error")
	}
}

func TestRabbitMQConnection_NewPublisher_Success(t *testing.T) {
	channel := &stubAMQPChannel{}
	stubConn := &stubAMQPConnection{ch: channel}
	rc := &RabbitMQConnection{conn: stubConn, logger: logger.NewNop()}

	if _, err := rc.NewPublisher("orders"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !channel.declareCalled {
		t.Fatal("expected exchange declare to be called")
	}
}

func TestRabbitMQConnection_NewPublisher_ChannelError(t *testing.T) {
	stubConn := &stubAMQPConnection{channelErr: errors.New("channel error")}
	rc := &RabbitMQConnection{conn: stubConn, logger: logger.NewNop()}

	if _, err := rc.NewPublisher("orders"); err == nil {
		t.Fatal("expected channel error")
	}
}

func TestRabbitMQConnection_CloseCallsUnderlying(t *testing.T) {
	stubConn := &stubAMQPConnection{}
	rc := &RabbitMQConnection{conn: stubConn}

	if err := rc.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
	if !stubConn.closed {
		t.Fatal("expected close to be forwarded")
	}
}

func TestNewRabbitMQConsumer_QosErrorClosesResources(t *testing.T) {
	channel := &stubAMQPChannel{qosErr: errors.New("qos error")}
	stubConn := &stubAMQPConnection{ch: channel}
	restore := SetAMQPDialer(func(url string) (AMQPConnection, error) {
		return stubConn, nil
	})
	defer restore()

	if _, err := NewRabbitMQConsumer("amqp://stub", "queue", logger.NewNop()); err == nil {
		t.Fatal("expected qos error")
	}
	if !channel.closed {
		t.Fatal("expected channel to close on qos error")
	}
	if !stubConn.closed {
		t.Fatal("expected connection to close on qos error")
	}
}

func TestNewRabbitMQConsumer_Success(t *testing.T) {
	channel := &stubAMQPChannel{}
	stubConn := &stubAMQPConnection{ch: channel}
	restore := SetAMQPDialer(func(url string) (AMQPConnection, error) {
		return stubConn, nil
	})
	defer restore()

	consumer, err := NewRabbitMQConsumer("amqp://stub", "queue", logger.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if consumer == nil {
		t.Fatal("expected consumer instance")
	}
	if !channel.qosCalled {
		t.Fatal("expected qos to be called")
	}
}

func TestNewRabbitMQConsumer_ChannelErrorClosesConnection(t *testing.T) {
	stubConn := &stubAMQPConnection{channelErr: errors.New("channel error")}
	restore := SetAMQPDialer(func(url string) (AMQPConnection, error) {
		return stubConn, nil
	})
	defer restore()

	if _, err := NewRabbitMQConsumer("amqp://stub", "queue", logger.NewNop()); err == nil {
		t.Fatal("expected channel error")
	}
	if !stubConn.closed {
		t.Fatal("expected connection to close when channel fails")
	}
}

func TestNewRabbitMQConsumer_DialError(t *testing.T) {
	restore := SetAMQPDialer(func(url string) (AMQPConnection, error) {
		return nil, errors.New("dial fail")
	})
	defer restore()

	if _, err := NewRabbitMQConsumer("amqp://stub", "queue", logger.NewNop()); err == nil {
		t.Fatal("expected dial error")
	}
}

func TestNewRabbitPublisher_DeclareErrorClosesResources(t *testing.T) {
	channel := &stubAMQPChannel{declareErr: errors.New("declare error")}
	stubConn := &stubAMQPConnection{ch: channel}
	restore := SetAMQPDialer(func(url string) (AMQPConnection, error) {
		return stubConn, nil
	})
	defer restore()

	if _, err := NewRabbitPublisher("amqp://stub", "orders", logger.NewNop()); err == nil {
		t.Fatal("expected declare error")
	}
	if !channel.closed {
		t.Fatal("expected channel to be closed on declare error")
	}
	if !stubConn.closed {
		t.Fatal("expected connection to be closed on declare error")
	}
}

func TestNewRabbitPublisher_Success(t *testing.T) {
	channel := &stubAMQPChannel{}
	stubConn := &stubAMQPConnection{ch: channel}
	restore := SetAMQPDialer(func(url string) (AMQPConnection, error) {
		return stubConn, nil
	})
	defer restore()

	pub, err := NewRabbitPublisher("amqp://stub", "orders", logger.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pub == nil {
		t.Fatal("expected publisher instance")
	}
	if !channel.declareCalled {
		t.Fatal("expected exchange declaration")
	}
}

func TestNewRabbitPublisher_DialError(t *testing.T) {
	restore := SetAMQPDialer(func(url string) (AMQPConnection, error) {
		return nil, errors.New("dial error")
	})
	defer restore()

	if _, err := NewRabbitPublisher("amqp://stub", "orders", logger.NewNop()); err == nil {
		t.Fatal("expected dial error")
	}
}

func TestNewRabbitMQConnection_DialError(t *testing.T) {
	restore := SetAMQPDialer(func(url string) (AMQPConnection, error) {
		return nil, errors.New("dial failed")
	})
	defer restore()

	if _, err := NewRabbitMQConnection("amqp://stub", logger.NewNop()); err == nil {
		t.Fatal("expected dial error")
	}
}
