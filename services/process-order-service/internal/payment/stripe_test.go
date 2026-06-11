package payment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fake Stripe API that records the idempotency key and returns a succeeded
// PaymentIntent.
func newFakeStripe(t *testing.T, status int, body map[string]any) (*httptest.Server, *http.Header) {
	t.Helper()
	var captured http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)
	return srv, &captured
}

func TestStripeProcessor_ChargeSendsIdempotencyKey(t *testing.T) {
	srv, captured := newFakeStripe(t, http.StatusOK, map[string]any{
		"id":     "pi_123",
		"status": "succeeded",
	})

	p := NewStripeProcessor("sk_test_123", WithBaseURL(srv.URL))

	if err := p.Charge(context.Background(), "order-42", 2500); err != nil {
		t.Fatalf("Charge: %v", err)
	}

	got := captured.Get("Idempotency-Key")
	if got != "order-42" {
		t.Fatalf("expected Idempotency-Key order-42, got %q", got)
	}
}

func TestStripeProcessor_CardDeclinedIsPermanent(t *testing.T) {
	srv, _ := newFakeStripe(t, http.StatusPaymentRequired, map[string]any{
		"error": map[string]any{
			"type":    "card_error",
			"code":    "card_declined",
			"message": "Your card was declined.",
		},
	})

	p := NewStripeProcessor("sk_test_123", WithBaseURL(srv.URL))

	err := p.Charge(context.Background(), "order-43", 1000)
	if err == nil {
		t.Fatal("expected error for declined card")
	}
	var permErr *PermanentError
	if !asPermanent(err, &permErr) {
		t.Fatalf("expected PermanentError, got %T: %v", err, err)
	}
}

func TestSimulatedProcessor_Succeeds(t *testing.T) {
	p := NewSimulatedProcessor(0) // no artificial latency in tests
	if err := p.Charge(context.Background(), "order-44", 500); err != nil {
		t.Fatalf("Charge: %v", err)
	}
}
