package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/icl00ud/publish-order-service/internal/model"
	"github.com/icl00ud/publish-order-service/internal/service"
)

type fakeRepo struct {
	saved model.Order
}

func (f *fakeRepo) Save(ctx context.Context, o model.Order) error {
	f.saved = o
	return nil
}
func (f *fakeRepo) Find(ctx context.Context, id string) (model.Order, error) {
	return f.saved, nil
}

type MockPub struct {
	Called bool
	Event  model.Event
}

func (m *MockPub) Publish(evt model.Event) error {
	m.Called = true
	m.Event = evt
	return nil
}

func TestCreateOrder_Success(t *testing.T) {
	repo := &fakeRepo{}
	pricing := service.NewPricingCalculator()
	svc := service.NewOrderService(repo, pricing)
	pub := &MockPub{}
	oh := NewOrderHandler(svc, pub)

	payload := `{"items":[{"product_id":"p1","name":"n1","quantity":2,"price":10.5}]}`
	req := httptest.NewRequest(http.MethodPost, "/create-order", strings.NewReader(payload))
	w := httptest.NewRecorder()

	oh.CreateOrder(w, req)

	if w.Result().StatusCode != http.StatusCreated {
		t.Fatalf("esperado 201, recebeu %d", w.Result().StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if !pub.Called {
		t.Error("esperava Publish ter sido chamado")
	}
	if pub.Event.Type != model.OrderCreated {
		t.Errorf("esperava evento %q, obteve %q", model.OrderCreated, pub.Event.Type)
	}
}
