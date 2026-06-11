package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/icl00ud/velure/services/process-order-service/internal/client"
	"github.com/icl00ud/velure/services/process-order-service/internal/model"
	"github.com/icl00ud/velure/services/process-order-service/internal/payment"
)

// Mock publisher for testing (thread-safe for parallel processing)
type mockPublisher struct {
	mu          sync.Mutex
	publishFunc func(evt model.Event) error
	published   []model.Event
}

func (m *mockPublisher) Publish(_ context.Context, evt model.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

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

// Mock product client for testing (thread-safe for parallel processing)
type mockProductClient struct {
	mu                 sync.Mutex
	updateQuantityFunc func(productID string, quantityChange int) error
	calls              []struct {
		productID      string
		quantityChange int
	}
}

func (m *mockProductClient) UpdateQuantity(_ context.Context, productID string, quantityChange int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

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

	svc := NewPaymentService(pub, client, &stubChargeProcessor{})

	if svc == nil {
		t.Fatal("expected non-nil payment service")
	}
}

func TestPaymentService_Process_Success(t *testing.T) {
	pub := &mockPublisher{}
	client := &mockProductClient{}

	svc := NewPaymentService(pub, client, &stubChargeProcessor{})

	items := []model.CartItem{
		{ProductID: "p1", Name: "Product 1", Quantity: 2, Price: 10.0},
		{ProductID: "p2", Name: "Product 2", Quantity: 1, Price: 5.0},
	}

	err := svc.Process(context.Background(), "order123", items, 25)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify product client was called for each item
	if len(client.calls) != 2 {
		t.Errorf("expected 2 UpdateQuantity calls, got %d", len(client.calls))
	}

	// Verify quantities were deducted (order not guaranteed due to parallel processing)
	callMap := make(map[string]int)
	for _, call := range client.calls {
		callMap[call.productID] = call.quantityChange
	}
	if callMap["p1"] != -2 {
		t.Errorf("expected UpdateQuantity(p1, -2), got quantityChange=%d", callMap["p1"])
	}
	if callMap["p2"] != -1 {
		t.Errorf("expected UpdateQuantity(p2, -1), got quantityChange=%d", callMap["p2"])
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

	svc := NewPaymentService(pub, client, &stubChargeProcessor{})

	items := []model.CartItem{
		{ProductID: "p1", Name: "Product 1", Quantity: 2, Price: 10.0},
	}

	err := svc.Process(context.Background(), "order123", items, 20)

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

	svc := NewPaymentService(pub, client, &stubChargeProcessor{})

	items := []model.CartItem{
		{ProductID: "p1", Name: "Product 1", Quantity: 1, Price: 10.0},
	}

	err := svc.Process(context.Background(), "order123", items, 10)

	if err == nil {
		t.Error("expected error, got nil")
	}

	// Inventory was deducted before the publish failure, then compensated back.
	if len(client.calls) != 2 {
		t.Errorf("expected 2 UpdateQuantity calls (deduct + compensate), got %d", len(client.calls))
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

	svc := NewPaymentService(pub, client, &stubChargeProcessor{})

	items := []model.CartItem{
		{ProductID: "p1", Name: "Product 1", Quantity: 1, Price: 10.0},
	}

	err := svc.Process(context.Background(), "order123", items, 10)

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

	svc := NewPaymentService(pub, client, &stubChargeProcessor{})

	items := []model.CartItem{
		{ProductID: "p1", Name: "Product 1", Quantity: 3, Price: 10.0},
		{ProductID: "p2", Name: "Product 2", Quantity: 2, Price: 15.0},
		{ProductID: "p3", Name: "Product 3", Quantity: 1, Price: 20.0},
	}

	err := svc.Process(context.Background(), "order456", items, 80)

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

	svc := NewPaymentService(pub, client, &stubChargeProcessor{})

	items := []model.CartItem{
		{ProductID: "p1", Name: "Product 1", Quantity: 1, Price: 10.0},
		{ProductID: "p2", Name: "Product 2", Quantity: 1, Price: 15.0},
		{ProductID: "p3", Name: "Product 3", Quantity: 1, Price: 20.0},
	}

	err := svc.Process(context.Background(), "order789", items, 45)

	if err == nil {
		t.Error("expected error, got nil")
	}

	// All 3 deductions run in parallel; the two that succeeded (p1, p3) are
	// compensated after p2 fails: 3 deductions + 2 compensations.
	if len(client.calls) != 5 {
		t.Errorf("expected 5 UpdateQuantity calls (3 deduct + 2 compensate), got %d", len(client.calls))
	}
}

func TestPaymentService_Process_PermanentErrorPublishesFailure(t *testing.T) {
	pub := &mockPublisher{}
	client := &mockProductClient{
		updateQuantityFunc: func(productID string, quantityChange int) error {
			return &client.PermanentError{Message: "not found", StatusCode: 404}
		},
	}

	svc := NewPaymentService(pub, client, &stubChargeProcessor{})

	items := []model.CartItem{
		{ProductID: "p1", Name: "Product 1", Quantity: 1, Price: 10.0},
	}

	err := svc.Process(context.Background(), "order999", items, 10)
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

	svc := NewPaymentService(pub, client, &stubChargeProcessor{})

	items := []model.CartItem{
		{ProductID: "p1", Name: "Product 1", Quantity: 1, Price: 10.0},
	}

	err := svc.Process(context.Background(), "order1000", items, 10)
	if err == nil {
		t.Fatal("expected error when publishing failure event fails")
	}
	if len(pub.published) != 1 {
		t.Fatalf("expected publish attempted once, got %d", len(pub.published))
	}
}

func TestPaymentService_Process_PartialFailureCompensatesDeductedItems(t *testing.T) {
	pub := &mockPublisher{}
	cli := &mockProductClient{
		updateQuantityFunc: func(productID string, quantityChange int) error {
			// Only the deduction of p2 fails; compensations (positive) succeed.
			if productID == "p2" && quantityChange < 0 {
				return errors.New("insufficient stock")
			}
			return nil
		},
	}

	svc := NewPaymentService(pub, cli, &stubChargeProcessor{})

	items := []model.CartItem{
		{ProductID: "p1", Name: "Product 1", Quantity: 2, Price: 10.0},
		{ProductID: "p2", Name: "Product 2", Quantity: 1, Price: 15.0},
		{ProductID: "p3", Name: "Product 3", Quantity: 4, Price: 20.0},
	}

	err := svc.Process(context.Background(), "order-comp-1", items, 105)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Net stock change per product must be zero for the succeeded deductions.
	net := map[string]int{}
	for _, call := range cli.calls {
		net[call.productID] += call.quantityChange
	}
	if net["p1"] != 0 {
		t.Errorf("expected p1 net change 0 (deduct then compensate), got %d", net["p1"])
	}
	if net["p3"] != 0 {
		t.Errorf("expected p3 net change 0 (deduct then compensate), got %d", net["p3"])
	}
}

func TestPaymentService_Process_PermanentErrorCompensatesDeductedItems(t *testing.T) {
	pub := &mockPublisher{}
	cli := &mockProductClient{
		updateQuantityFunc: func(productID string, quantityChange int) error {
			if productID == "p2" && quantityChange < 0 {
				return &client.PermanentError{Message: "not found", StatusCode: 404}
			}
			return nil
		},
	}

	svc := NewPaymentService(pub, cli, &stubChargeProcessor{})

	items := []model.CartItem{
		{ProductID: "p1", Name: "Product 1", Quantity: 1, Price: 10.0},
		{ProductID: "p2", Name: "Product 2", Quantity: 1, Price: 15.0},
	}

	err := svc.Process(context.Background(), "order-comp-2", items, 25)
	if err != nil {
		t.Fatalf("expected nil error for permanent failure (ack), got %v", err)
	}

	net := map[string]int{}
	for _, call := range cli.calls {
		net[call.productID] += call.quantityChange
	}
	if net["p1"] != 0 {
		t.Errorf("expected p1 net change 0 after compensation, got %d", net["p1"])
	}

	if len(pub.published) != 1 || pub.published[0].Type != model.OrderFailed {
		t.Fatalf("expected one OrderFailed event, got %+v", pub.published)
	}
}

func TestPaymentService_Process_PublishProcessingFailureCompensates(t *testing.T) {
	pub := &mockPublisher{
		publishFunc: func(evt model.Event) error {
			if evt.Type == model.OrderProcessing {
				return errors.New("rabbitmq down")
			}
			return nil
		},
	}
	cli := &mockProductClient{}

	svc := NewPaymentService(pub, cli, &stubChargeProcessor{})

	items := []model.CartItem{
		{ProductID: "p1", Name: "Product 1", Quantity: 3, Price: 10.0},
	}

	err := svc.Process(context.Background(), "order-comp-3", items, 30)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Error return means the message will be retried and stock re-deducted,
	// so the failed attempt must leave net stock unchanged.
	net := 0
	for _, call := range cli.calls {
		net += call.quantityChange
	}
	if net != 0 {
		t.Errorf("expected net stock change 0 before retry, got %d", net)
	}
}

type stubChargeProcessor struct {
	err   error
	calls []string
}

func (s *stubChargeProcessor) Charge(_ context.Context, orderID string, amountCents int64) error {
	s.calls = append(s.calls, orderID)
	return s.err
}

func TestPaymentService_Process_ChargesViaProcessor(t *testing.T) {
	pub := &mockPublisher{}
	cli := &mockProductClient{}
	proc := &stubChargeProcessor{}

	svc := NewPaymentService(pub, cli, proc)

	items := []model.CartItem{{ProductID: "p1", Name: "P1", Quantity: 1, Price: 10.0}}
	if err := svc.Process(context.Background(), "order-pay-1", items, 1000); err != nil {
		t.Fatalf("Process: %v", err)
	}
	if len(proc.calls) != 1 || proc.calls[0] != "order-pay-1" {
		t.Fatalf("expected processor charged once for order-pay-1, got %v", proc.calls)
	}
}

func TestPaymentService_Process_PermanentChargeFailureFailsOrder(t *testing.T) {
	pub := &mockPublisher{}
	cli := &mockProductClient{}
	proc := &stubChargeProcessor{err: &payment.PermanentError{Reason: "card declined"}}

	svc := NewPaymentService(pub, cli, proc)

	items := []model.CartItem{{ProductID: "p1", Name: "P1", Quantity: 2, Price: 10.0}}
	err := svc.Process(context.Background(), "order-pay-2", items, 2000)
	if err != nil {
		t.Fatalf("permanent decline should ack (nil error), got %v", err)
	}

	// Stock handed back.
	net := 0
	for _, call := range cli.calls {
		net += call.quantityChange
	}
	if net != 0 {
		t.Fatalf("expected net stock change 0 after declined payment, got %d", net)
	}

	// OrderFailed published (after the OrderProcessing event).
	last := pub.published[len(pub.published)-1]
	if last.Type != model.OrderFailed {
		t.Fatalf("expected final event OrderFailed, got %s", last.Type)
	}
}

func TestPaymentService_Process_TransientChargeFailureRetries(t *testing.T) {
	pub := &mockPublisher{}
	cli := &mockProductClient{}
	proc := &stubChargeProcessor{err: errors.New("stripe 502")}

	svc := NewPaymentService(pub, cli, proc)

	items := []model.CartItem{{ProductID: "p1", Name: "P1", Quantity: 1, Price: 10.0}}
	err := svc.Process(context.Background(), "order-pay-3", items, 1000)
	if err == nil {
		t.Fatal("transient charge failure must return error for retry")
	}

	net := 0
	for _, call := range cli.calls {
		net += call.quantityChange
	}
	if net != 0 {
		t.Fatalf("expected net stock change 0 before retry, got %d", net)
	}
}
