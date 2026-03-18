package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewProductClient(t *testing.T) {
	client := NewProductClient("http://example.com")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestProductClient_UpdateQuantity_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH request, got %s", r.Method)
		}
		if r.URL.Path != "/api/products/product123/inventory" {
			t.Errorf("expected path /api/products/product123/inventory, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed reading body: %v", err)
		}
		var payload map[string]int
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("failed unmarshalling body: %v", err)
		}
		if _, exists := payload["product_id"]; exists {
			t.Errorf("product_id should not be sent in body")
		}
		if payload["quantity_change"] != -2 {
			t.Errorf("expected quantity_change -2, got %d", payload["quantity_change"])
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewProductClient(server.URL)
	err := client.UpdateQuantity("product123", -2)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProductClient_UpdateQuantity_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"insufficient stock"}`))
	}))
	defer server.Close()

	client := NewProductClient(server.URL)
	err := client.UpdateQuantity("product123", -10)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if _, ok := err.(*PermanentError); !ok {
		t.Errorf("expected PermanentError, got %T: %v", err, err)
	}
}

func TestProductClient_UpdateQuantity_NonJSONError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`Internal Server Error`))
	}))
	defer server.Close()

	client := NewProductClient(server.URL)
	err := client.UpdateQuantity("product123", -2)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if _, ok := err.(*TransientError); !ok {
		t.Errorf("expected TransientError, got %T: %v", err, err)
	}
}

func TestProductClient_UpdateQuantity_InvalidURL(t *testing.T) {
	client := NewProductClient("http://invalid-url-that-does-not-exist.local:99999")
	err := client.UpdateQuantity("product123", -2)

	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestProductClient_UpdateQuantity_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewProductClient(server.URL)
	err := client.UpdateQuantity("product123", -2)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProductClient_UpdateQuantity_PositiveChange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewProductClient(server.URL)
	err := client.UpdateQuantity("product456", 5)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProductClient_UpdateQuantity_TooManyRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":"retry later"}`))
	}))
	defer server.Close()

	client := NewProductClient(server.URL)
	err := client.UpdateQuantity("product789", -1)
	if err == nil {
		t.Fatal("expected transient error")
	}
	if te, ok := err.(*TransientError); !ok || te.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected TransientError with status 429, got %T: %v", err, err)
	}
}

func TestProductClient_UpdateQuantity_UnexpectedStatusDefaultsToPermanent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	defer server.Close()

	client := NewProductClient(server.URL)
	err := client.UpdateQuantity("product123", -1)
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(*PermanentError); !ok {
		t.Fatalf("expected PermanentError for unexpected status, got %T", err)
	}
}

func TestPermanentAndTransientErrorMessages(t *testing.T) {
	pe := &PermanentError{Message: "not found", StatusCode: http.StatusNotFound}
	if pe.Error() == "" {
		t.Fatal("expected permanent error message")
	}

	te := &TransientError{Message: "try again", StatusCode: http.StatusTooManyRequests}
	if te.Error() == "" {
		t.Fatal("expected transient error message")
	}
}
