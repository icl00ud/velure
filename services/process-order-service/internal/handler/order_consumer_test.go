package handler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/icl00ud/process-order-service/internal/model"
	"github.com/icl00ud/velure-shared/logger"
)

// Mock consumer for testing
type mockConsumer struct {
	consumeFunc func(ctx context.Context, handler func(model.Event) error) error
}

func (m *mockConsumer) Consume(ctx context.Context, handler func(model.Event) error) error {
	if m.consumeFunc != nil {
		return m.consumeFunc(ctx, handler)
	}
	return nil
}

func (m *mockConsumer) Close() error {
	return nil
}

// Mock payment service for testing
type mockPaymentService struct {
	processFunc func(orderID string, items []model.CartItem, amount int) error
	calls       []struct {
		orderID string
		items   []model.CartItem
		amount  int
	}
}

func (m *mockPaymentService) Process(orderID string, items []model.CartItem, amount int) error {
	if m.calls == nil {
		m.calls = []struct {
			orderID string
			items   []model.CartItem
			amount  int
		}{}
	}
	m.calls = append(m.calls, struct {
		orderID string
		items   []model.CartItem
		amount  int
	}{orderID, items, amount})

	if m.processFunc != nil {
		return m.processFunc(orderID, items, amount)
	}
	return nil
}

func TestNewOrderConsumer(t *testing.T) {
	consumer := &mockConsumer{}
	svc := &mockPaymentService{}
	logger := logger.NewNop()

	oc := NewOrderConsumer(consumer, svc, 5, logger)

	if oc == nil {
		t.Fatal("expected non-nil order consumer")
	}
	if oc.workers != 5 {
		t.Errorf("expected workers 5, got %d", oc.workers)
	}
}

func TestOrderConsumer_Start_OrderCreatedEvent(t *testing.T) {
	svc := &mockPaymentService{}
	logger := logger.NewNop()

	// Create a test event
	orderPayload := struct {
		ID    string           `json:"id"`
		Items []model.CartItem `json:"items"`
		Total float64          `json:"total"`
	}{
		ID: "order123",
		Items: []model.CartItem{
			{ProductID: "p1", Name: "Product 1", Quantity: 2, Price: 10.0},
		},
		Total: 20.0,
	}
	payloadBytes, _ := json.Marshal(orderPayload)

	evt := model.Event{
		Type:    model.OrderCreated,
		Payload: payloadBytes,
	}

	// Create consumer that immediately calls handler with test event
	consumer := &mockConsumer{
		consumeFunc: func(ctx context.Context, handler func(model.Event) error) error {
			// Call handler with test event
			return handler(evt)
		},
	}

	oc := NewOrderConsumer(consumer, svc, 1, logger)

	// Start consumer in a goroutine
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start consumer
	done := make(chan error, 1)
	go func() {
		done <- oc.Start(ctx)
	}()

	// Wait for context to be done
	<-ctx.Done()

	// Verify payment service was called
	if len(svc.calls) != 1 {
		t.Errorf("expected 1 Process call, got %d", len(svc.calls))
	}
	if len(svc.calls) > 0 {
		if svc.calls[0].orderID != "order123" {
			t.Errorf("expected orderID order123, got %s", svc.calls[0].orderID)
		}
		if svc.calls[0].amount != 20 {
			t.Errorf("expected amount 20, got %d", svc.calls[0].amount)
		}
	}
}

func TestOrderConsumer_Start_NonOrderCreatedEvent(t *testing.T) {
	svc := &mockPaymentService{}
	logger := logger.NewNop()

	// Create a non-order.created event
	evt := model.Event{
		Type:    model.OrderProcessing,
		Payload: json.RawMessage(`{"id":"order123"}`),
	}

	consumer := &mockConsumer{
		consumeFunc: func(ctx context.Context, handler func(model.Event) error) error {
			return handler(evt)
		},
	}

	oc := NewOrderConsumer(consumer, svc, 1, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		oc.Start(ctx)
	}()

	<-ctx.Done()

	// Verify payment service was NOT called
	if len(svc.calls) != 0 {
		t.Errorf("expected 0 Process calls, got %d", len(svc.calls))
	}
}

func TestOrderConsumer_Start_InvalidPayload(t *testing.T) {
	svc := &mockPaymentService{}
	logger := logger.NewNop()

	// Create event with invalid JSON payload
	evt := model.Event{
		Type:    model.OrderCreated,
		Payload: json.RawMessage(`{invalid json`),
	}

	consumer := &mockConsumer{
		consumeFunc: func(ctx context.Context, handler func(model.Event) error) error {
			return handler(evt)
		},
	}

	oc := NewOrderConsumer(consumer, svc, 1, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		oc.Start(ctx)
	}()

	<-ctx.Done()

	// Verify payment service was NOT called (due to unmarshal error)
	if len(svc.calls) != 0 {
		t.Errorf("expected 0 Process calls, got %d", len(svc.calls))
	}
}

func TestOrderConsumer_Start_ProcessFails(t *testing.T) {
	svc := &mockPaymentService{
		processFunc: func(orderID string, items []model.CartItem, amount int) error {
			return errors.New("payment processing failed")
		},
	}
	logger := logger.NewNop()

	orderPayload := struct {
		ID    string           `json:"id"`
		Items []model.CartItem `json:"items"`
		Total float64          `json:"total"`
	}{
		ID: "order123",
		Items: []model.CartItem{
			{ProductID: "p1", Name: "Product 1", Quantity: 1, Price: 10.0},
		},
		Total: 10.0,
	}
	payloadBytes, _ := json.Marshal(orderPayload)

	evt := model.Event{
		Type:    model.OrderCreated,
		Payload: payloadBytes,
	}

	consumer := &mockConsumer{
		consumeFunc: func(ctx context.Context, handler func(model.Event) error) error {
			return handler(evt)
		},
	}

	oc := NewOrderConsumer(consumer, svc, 1, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		oc.Start(ctx)
	}()

	<-ctx.Done()

	// Verify payment service was called even though it failed
	if len(svc.calls) != 1 {
		t.Errorf("expected 1 Process call, got %d", len(svc.calls))
	}
}

func TestOrderConsumer_Start_MultipleWorkers(t *testing.T) {
	svc := &mockPaymentService{}
	logger := logger.NewNop()

	consumer := &mockConsumer{
		consumeFunc: func(ctx context.Context, handler func(model.Event) error) error {
			// Wait for context to be done
			<-ctx.Done()
			return nil
		},
	}

	oc := NewOrderConsumer(consumer, svc, 10, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := oc.Start(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
