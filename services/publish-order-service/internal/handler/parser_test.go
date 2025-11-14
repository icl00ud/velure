package handler

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseCreateOrder_Valid(t *testing.T) {
	input := `{"items":[{"product_id":"p1","name":"n1","quantity":2,"price":5.5}]}`
	items, err := parseCreateOrder(strings.NewReader(input))
	if err != nil {
		t.Fatalf("esperava sem erro, teve: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("esperava 1 item, obteve %d", len(items))
	}
	if items[0].ProductID != "p1" || items[0].Name != "n1" || items[0].Quantity != 2 || items[0].Price != 5.5 {
		t.Errorf("item incorreto: %+v", items[0])
	}
}

func TestParseCreateOrder_Invalid(t *testing.T) {
	_, err := parseCreateOrder(strings.NewReader("%%%"))
	if err == nil {
		t.Fatal("esperava erro de JSON inválido, mas err==nil")
	}
}

func TestParseCreateOrder_DirectArray(t *testing.T) {
	input := `[{"product_id":"p2","name":"Product 2","quantity":3,"price":15.0}]`
	items, err := parseCreateOrder(strings.NewReader(input))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ProductID != "p2" {
		t.Errorf("expected product_id p2, got %s", items[0].ProductID)
	}
}

func TestParsePagination_Defaults(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	page, pageSize := parsePagination(req)

	if page != 1 {
		t.Errorf("expected default page 1, got %d", page)
	}
	if pageSize != 10 {
		t.Errorf("expected default pageSize 10, got %d", pageSize)
	}
}

func TestParsePagination_CustomValues(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?page=5&pageSize=25", nil)
	page, pageSize := parsePagination(req)

	if page != 5 {
		t.Errorf("expected page 5, got %d", page)
	}
	if pageSize != 25 {
		t.Errorf("expected pageSize 25, got %d", pageSize)
	}
}

func TestParsePagination_InvalidPage(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?page=invalid", nil)
	page, pageSize := parsePagination(req)

	if page != 1 {
		t.Errorf("expected default page 1 for invalid input, got %d", page)
	}
	if pageSize != 10 {
		t.Errorf("expected default pageSize 10, got %d", pageSize)
	}
}

func TestParsePagination_ZeroAndNegative(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?page=0&pageSize=-5", nil)
	page, pageSize := parsePagination(req)

	if page != 1 {
		t.Errorf("expected default page 1 for zero/negative, got %d", page)
	}
	if pageSize != 10 {
		t.Errorf("expected default pageSize 10 for zero/negative, got %d", pageSize)
	}
}

func TestParsePagination_ExceedsMax(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?page=2&pageSize=200", nil)
	page, pageSize := parsePagination(req)

	if page != 2 {
		t.Errorf("expected page 2, got %d", page)
	}
	if pageSize != 10 {
		t.Errorf("expected default pageSize 10 for exceeding max, got %d", pageSize)
	}
}
