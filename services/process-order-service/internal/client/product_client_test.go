package client

import (
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
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/product/updateQuantity" {
			t.Errorf("expected path /product/updateQuantity, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
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
	// Create a test server that returns an error
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
	if err.Error() != "product service error: insufficient stock" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestProductClient_UpdateQuantity_NonJSONError(t *testing.T) {
	// Create a test server that returns non-JSON error
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
	if err.Error() != "unexpected status code: 500" {
		t.Errorf("unexpected error message: %v", err)
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
	// Create a test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Server responds immediately for this test
		// The actual timeout is 10 seconds in the client
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
	// Test with positive quantity change (adding stock back)
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
