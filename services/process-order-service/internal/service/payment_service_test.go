package service

import (
	"errors"
	"testing"

	"github.com/icl00ud/process-order-service/internal/client"
	"github.com/icl00ud/process-order-service/internal/model"
)

// Mock publisher for testing
type mockPublisher struct {
	publishFunc func(evt model.Event) error
	published   []model.Event
}

func (m *mockPublisher) Publish(evt model.Event) error {
	if m.published == nil {
		m.published = []model.Event{}
	}
	m.published = append(m.published, evt)

	if m.publishFunc != nil {
		return m.publishFunc(evt)
	}
	return nil
}

func (m *mockPublisher) Close() error {
	return nil
}

// Mock product client for testing
type mockProductClient struct {
	updateQuantityFunc func(productID string, quantityChange int) error
	calls              []struct {
		productID      string
		quantityChange int
	}
}

func (m *mockProductClient) UpdateQuantity(productID string, quantityChange int) error {
	if m.calls == nil {
		m.calls = []struct {
			productID      string
			quantityChange int
		}{}
	}
	m.calls = append(m.calls, struct {
		productID      string
		quantityChange int
	}{productID, quantityChange})

	if m.updateQuantityFunc != nil {
		return m.updateQuantityFunc(productID, quantityChange)
	}
	return nil
}

func TestNewPaymentService(t *testing.T) {
	pub := &mockPublisher{}
	client := &mockProductClient{}

	svc := NewPaymentService(pub, client)

	if svc == nil {
		t.Fatal("expected non-nil payment service")
	}
}

func TestPaymentService_Process_Success(t *testing.T) {
	pub := &mockPublisher{}
	client := &mockProductClient{}

	svc := NewPaymentService(pub, client)

	items := []model.CartItem{
		{ProductID: "p1", Name: "Product 1", Quantity: 2, Price: 10.0},
		{ProductID: "p2", Name: "Product 2", Quantity: 1, Price: 5.0},
	}

	err := svc.Process("order123", items, 25)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify product client was called for each item
	if len(client.calls) != 2 {
		t.Errorf("expected 2 UpdateQuantity calls, got %d", len(client.calls))
	}

	// Verify quantities were deducted
	if client.calls[0].productID != "p1" || client.calls[0].quantityChange != -2 {
		t.Errorf("expected UpdateQuantity(p1, -2), got UpdateQuantity(%s, %d)",
			client.calls[0].productID, client.calls[0].quantityChange)
	}
	if client.calls[1].productID != "p2" || client.calls[1].quantityChange != -1 {
		t.Errorf("expected UpdateQuantity(p2, -1), got UpdateQuantity(%s, %d)",
			client.calls[1].productID, client.calls[1].quantityChange)
	}

	// Verify events were published (processing and completed)
	if len(pub.published) != 2 {
		t.Errorf("expected 2 events published, got %d", len(pub.published))
	}

	if pub.published[0].Type != model.OrderProcessing {
		t.Errorf("expected first event type %s, got %s", model.OrderProcessing, pub.published[0].Type)
	}

	if pub.published[1].Type != model.OrderCompleted {
		t.Errorf("expected second event type %s, got %s", model.OrderCompleted, pub.published[1].Type)
	}
}

func TestPaymentService_Process_InventoryUpdateFails(t *testing.T) {
	pub := &mockPublisher{}
	client := &mockProductClient{
		updateQuantityFunc: func(productID string, quantityChange int) error {
			return errors.New("insufficient stock")
		},
	}

	svc := NewPaymentService(pub, client)

	items := []model.CartItem{
		{ProductID: "p1", Name: "Product 1", Quantity: 2, Price: 10.0},
	}

	err := svc.Process("order123", items, 20)

	if err == nil {
		t.Error("expected error, got nil")
	}

	// Verify no events were published (since inventory update failed)
	if len(pub.published) != 0 {
		t.Errorf("expected 0 events published, got %d", len(pub.published))
	}
}

func TestPaymentService_Process_PublishProcessingFails(t *testing.T) {
	pub := &mockPublisher{
		publishFunc: func(evt model.Event) error {
			if evt.Type == model.OrderProcessing {
				return errors.New("rabbitmq connection failed")
			}
			return nil
		},
	}
	client := &mockProductClient{}

	svc := NewPaymentService(pub, client)

	items := []model.CartItem{
		{ProductID: "p1", Name: "Product 1", Quantity: 1, Price: 10.0},
	}

	err := svc.Process("order123", items, 10)

	if err == nil {
		t.Error("expected error, got nil")
	}

	// Verify inventory was still deducted (since it happens before publish)
	if len(client.calls) != 1 {
		t.Errorf("expected 1 UpdateQuantity call, got %d", len(client.calls))
	}
}

func TestPaymentService_Process_PublishCompletedFails(t *testing.T) {
	pub := &mockPublisher{
		publishFunc: func(evt model.Event) error {
			if evt.Type == model.OrderCompleted {
				return errors.New("rabbitmq connection failed")
			}
			return nil
		},
	}
	client := &mockProductClient{}

	svc := NewPaymentService(pub, client)

	items := []model.CartItem{
		{ProductID: "p1", Name: "Product 1", Quantity: 1, Price: 10.0},
	}

	err := svc.Process("order123", items, 10)

	if err == nil {
		t.Error("expected error, got nil")
	}

	// Verify both events were attempted to be published (processing succeeded, completed failed)
	if len(pub.published) != 2 {
		t.Errorf("expected 2 events published, got %d", len(pub.published))
	}
}

func TestPaymentService_Process_MultipleItems(t *testing.T) {
	pub := &mockPublisher{}
	client := &mockProductClient{}

	svc := NewPaymentService(pub, client)

	items := []model.CartItem{
		{ProductID: "p1", Name: "Product 1", Quantity: 3, Price: 10.0},
		{ProductID: "p2", Name: "Product 2", Quantity: 2, Price: 15.0},
		{ProductID: "p3", Name: "Product 3", Quantity: 1, Price: 20.0},
	}

	err := svc.Process("order456", items, 80)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify all items had inventory updated
	if len(client.calls) != 3 {
		t.Errorf("expected 3 UpdateQuantity calls, got %d", len(client.calls))
	}
}

func TestPaymentService_Process_SecondItemFails(t *testing.T) {
	pub := &mockPublisher{}
	client := &mockProductClient{
		updateQuantityFunc: func(productID string, quantityChange int) error {
			if productID == "p2" {
				return errors.New("product not found")
			}
			return nil
		},
	}

	svc := NewPaymentService(pub, client)

	items := []model.CartItem{
		{ProductID: "p1", Name: "Product 1", Quantity: 1, Price: 10.0},
		{ProductID: "p2", Name: "Product 2", Quantity: 1, Price: 15.0},
		{ProductID: "p3", Name: "Product 3", Quantity: 1, Price: 20.0},
	}

	err := svc.Process("order789", items, 45)

	if err == nil {
		t.Error("expected error, got nil")
	}

	// First item should have been processed, but process stopped at second item
	if len(client.calls) != 2 {
		t.Errorf("expected 2 UpdateQuantity calls, got %d", len(client.calls))
	}
}

func TestPaymentService_Process_PermanentErrorPublishesFailure(t *testing.T) {
	pub := &mockPublisher{}
	client := &mockProductClient{
		updateQuantityFunc: func(productID string, quantityChange int) error {
			return &client.PermanentError{Message: "not found", StatusCode: 404}
		},
	}

	svc := NewPaymentService(pub, client)

	items := []model.CartItem{
		{ProductID: "p1", Name: "Product 1", Quantity: 1, Price: 10.0},
	}

	err := svc.Process("order999", items, 10)
	if err != nil {
		t.Fatalf("expected nil error for permanent errors (ack), got %v", err)
	}

	if len(pub.published) != 1 {
		t.Fatalf("expected failure event published, got %d", len(pub.published))
	}
	if pub.published[0].Type != model.OrderFailed {
		t.Fatalf("expected event type %s, got %s", model.OrderFailed, pub.published[0].Type)
	}
}

func TestPaymentService_Process_PermanentErrorPublishFails(t *testing.T) {
	pub := &mockPublisher{
		publishFunc: func(evt model.Event) error {
			return errors.New("publish fail")
		},
	}
	client := &mockProductClient{
		updateQuantityFunc: func(productID string, quantityChange int) error {
			return &client.PermanentError{Message: "not found", StatusCode: 404}
		},
	}

	svc := NewPaymentService(pub, client)

	items := []model.CartItem{
		{ProductID: "p1", Name: "Product 1", Quantity: 1, Price: 10.0},
	}

	err := svc.Process("order1000", items, 10)
	if err == nil {
		t.Fatal("expected error when publishing failure event fails")
	}
	if len(pub.published) != 1 {
		t.Fatalf("expected publish attempted once, got %d", len(pub.published))
	}
}
