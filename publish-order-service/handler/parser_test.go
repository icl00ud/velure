package handler

import (
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
