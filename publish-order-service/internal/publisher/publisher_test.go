package publisher

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/icl00ud/publish-order-service/internal/model"
	"github.com/rabbitmq/amqp091-go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

func setupRabbitContainer(ctx context.Context, t *testing.T) (testcontainers.Container, string) {
	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3.11-management",
		ExposedPorts: []string{"5672/tcp"},
		WaitingFor:   wait.ForLog("Server startup complete"),
	}
	rabbitC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start RabbitMQ container: %v", err)
	}
	host, _ := rabbitC.Host(ctx)
	port, _ := rabbitC.MappedPort(ctx, "5672")
	url := "amqp://guest:guest@" + host + ":" + port.Port() + "/"
	return rabbitC, url
}

func TestNewRabbitMQPublisher_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	rabbitC, url := setupRabbitContainer(ctx, t)
	defer rabbitC.Terminate(ctx)
	logger := zap.NewNop()

	// Act
	pub, err := NewRabbitMQPublisher(url, "ex.test", logger)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer pub.Close()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestNewRabbitMQPublisher_DialError(t *testing.T) {
	// Arrange
	invalidURL := "amqp://invalid:invalid/"

	// Act
	_, err := NewRabbitMQPublisher(invalidURL, "ex.test", zap.NewNop())

	// Assert
	if err == nil {
		t.Fatal("expected dial error, got nil")
	}
}

func TestNewRabbitMQPublisher_ExchangeDeclareError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	rabbitC, url := setupRabbitContainer(ctx, t)
	defer rabbitC.Terminate(ctx)

	// Act
	_, err := NewRabbitMQPublisher(url, "", zap.NewNop())

	// Assert
	if err == nil {
		t.Fatal("expected exchange declare error, got nil")
	}
}

func TestPublishAndConsume(t *testing.T) {
	// Arrange
	ctx := context.Background()
	rabbitC, url := setupRabbitContainer(ctx, t)
	defer rabbitC.Terminate(ctx)

	logger := zap.NewNop()
	exchange := "ex.test"
	routingKey := "test.key"

	pub, err := NewRabbitMQPublisher(url, exchange, logger)
	if err != nil {
		t.Fatalf("setup publisher failed: %v", err)
	}
	defer pub.Close()

	conn, err := amqp091.Dial(url)
	if err != nil {
		t.Fatalf("consumer dial failed: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("consumer channel failed: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		t.Fatalf("queue declare failed: %v", err)
	}
	if err := ch.QueueBind(q.Name, routingKey, exchange, false, nil); err != nil {
		t.Fatalf("queue bind failed: %v", err)
	}

	payload := map[string]interface{}{"foo": "bar"}
	body, _ := json.Marshal(payload)
	evt := model.Event{Type: routingKey, Payload: body}

	// Act
	if err := pub.Publish(evt); err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	msgs, err := ch.Consume(q.Name, "", true, true, false, false, nil)
	if err != nil {
		t.Fatalf("consume failed: %v", err)
	}

	// Assert
	select {
	case msg := <-msgs:
		var got model.Event
		if err := json.Unmarshal(msg.Body, &got); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if got.Type != evt.Type {
			t.Errorf("routing key mismatch: want %s, got %s", evt.Type, got.Type)
		}
		var payloadGot map[string]interface{}
		if err := json.Unmarshal(got.Payload, &payloadGot); err != nil {
			t.Fatalf("payload unmarshal failed: %v", err)
		}
		if payloadGot["foo"] != "bar" {
			t.Errorf("payload mismatch: want foo=bar, got %v", payloadGot)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestPublish_AfterClose(t *testing.T) {
	// Arrange
	ctx := context.Background()
	rabbitC, url := setupRabbitContainer(ctx, t)
	defer rabbitC.Terminate(ctx)

	pub, err := NewRabbitMQPublisher(url, "ex.test", zap.NewNop())
	if err != nil {
		t.Fatalf("setup publisher failed: %v", err)
	}

	if err := pub.Close(); err != nil {
		t.Fatalf("first close failed: %v", err)
	}

	evt := model.Event{Type: "x", Payload: []byte(`{}`)}

	// Act
	done := make(chan error, 1)
	go func() {
		done <- pub.Publish(evt)
	}()

	// Assert
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected error when publishing after close, got nil")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("publish after close hung")
	}
}

func TestClose_Idempotent(t *testing.T) {
	// Arrange
	ctx := context.Background()
	rabbitC, url := setupRabbitContainer(ctx, t)
	defer rabbitC.Terminate(ctx)

	pub, err := NewRabbitMQPublisher(url, "ex.test", zap.NewNop())
	if err != nil {
		t.Fatalf("setup publisher failed: %v", err)
	}

	// Act & Assert
	if err := pub.Close(); err != nil {
		t.Errorf("first close error: %v", err)
	}
	if err := pub.Close(); err != nil {
		t.Errorf("second close error: %v", err)
	}
}
