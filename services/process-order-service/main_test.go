package main

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"github.com/icl00ud/process-order-service/internal/queue"
)

type testChannel struct {
	consumeCh    <-chan amqp091.Delivery
	qosCalls     int
	bindCalls    int
	declareCalls int
	closed       bool
	consumeErr   error
	declareErr   error
	queueBindErr error
	qosErr       error
}

func (t *testChannel) Consume(queueName, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error) {
	if t.consumeErr != nil {
		return nil, t.consumeErr
	}
	if t.consumeCh == nil {
		ch := make(chan amqp091.Delivery)
		close(ch)
		t.consumeCh = ch
	}
	return t.consumeCh, nil
}

func (t *testChannel) Qos(prefetchCount, prefetchSize int, global bool) error {
	t.qosCalls++
	return t.qosErr
}

func (t *testChannel) QueueBind(queueName, key, exchange string, noWait bool, args amqp091.Table) error {
	t.bindCalls++
	return t.queueBindErr
}

func (t *testChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
	return nil
}

func (t *testChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp091.Table) error {
	t.declareCalls++
	return t.declareErr
}

func (t *testChannel) Close() error {
	t.closed = true
	return nil
}

type sequencedConnection struct {
	channels []queue.AMQPChannel
	closed   bool
}

func (c *sequencedConnection) Channel() (queue.AMQPChannel, error) {
	if len(c.channels) == 0 {
		return nil, errors.New("no channel available")
	}
	ch := c.channels[0]
	c.channels = c.channels[1:]
	return ch, nil
}

func (c *sequencedConnection) Close() error {
	c.closed = true
	return nil
}

func TestRun_StartsAndStopsWithStubbedDependencies(t *testing.T) {
	t.Setenv("PROCESS_ORDER_SERVICE_APP_PORT", "0")
	t.Setenv("PROCESS_RABBITMQ_URL", "amqp://stub")
	t.Setenv("RABBITMQ_ORDER_QUEUE", "orders")
	t.Setenv("ORDER_EXCHANGE", "orders")
	t.Setenv("WORKERS", "1")

	consumerChannel := &testChannel{}
	publisherChannel := &testChannel{}
	conn := &sequencedConnection{channels: []queue.AMQPChannel{consumerChannel, publisherChannel}}

	restore := queue.SetAMQPDialer(func(url string) (queue.AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, zap.NewNop())
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for run to return")
	}

	if !conn.closed {
		t.Fatal("expected connection to be closed")
	}
	if !consumerChannel.closed {
		t.Fatal("expected consumer channel to be closed")
	}
	if !publisherChannel.closed {
		t.Fatal("expected publisher channel to be closed")
	}
}

func TestMain_ExitsOnSignal(t *testing.T) {
	t.Setenv("PROCESS_ORDER_SERVICE_APP_PORT", "0")
	t.Setenv("PROCESS_RABBITMQ_URL", "amqp://stub")
	t.Setenv("RABBITMQ_ORDER_QUEUE", "orders")
	t.Setenv("ORDER_EXCHANGE", "orders")
	t.Setenv("WORKERS", "1")

	consumerChannel := &testChannel{}
	publisherChannel := &testChannel{}
	conn := &sequencedConnection{channels: []queue.AMQPChannel{consumerChannel, publisherChannel}}

	restore := queue.SetAMQPDialer(func(url string) (queue.AMQPConnection, error) {
		return conn, nil
	})
	defer restore()

	done := make(chan struct{})
	go func() {
		main()
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatalf("failed to signal process: %v", err)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("main did not exit on signal")
	}
}
