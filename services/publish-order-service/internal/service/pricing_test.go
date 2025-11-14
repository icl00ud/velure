package service

import (
	"testing"

	"github.com/icl00ud/publish-order-service/internal/model"
)

func TestNewPricingCalculator(t *testing.T) {
	pc := NewPricingCalculator()
	if pc == nil {
		t.Fatal("expected non-nil pricing calculator")
	}
}

func TestDefaultPricing_Calculate(t *testing.T) {
	tests := []struct {
		name     string
		items    []model.CartItem
		expected float64
	}{
		{
			name:     "empty items",
			items:    []model.CartItem{},
			expected: 0.0,
		},
		{
			name: "single item",
			items: []model.CartItem{
				{ProductID: "p1", Name: "Product 1", Price: 10.0, Quantity: 2},
			},
			expected: 20.0,
		},
		{
			name: "multiple items",
			items: []model.CartItem{
				{ProductID: "p1", Name: "Product 1", Price: 10.0, Quantity: 2},
				{ProductID: "p2", Name: "Product 2", Price: 5.5, Quantity: 3},
			},
			expected: 36.5,
		},
		{
			name: "items with decimal prices",
			items: []model.CartItem{
				{ProductID: "p1", Name: "Product 1", Price: 9.99, Quantity: 1},
				{ProductID: "p2", Name: "Product 2", Price: 19.99, Quantity: 2},
			},
			expected: 49.97,
		},
		{
			name: "zero price items",
			items: []model.CartItem{
				{ProductID: "p1", Name: "Free Item", Price: 0.0, Quantity: 5},
			},
			expected: 0.0,
		},
	}

	pc := NewPricingCalculator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pc.Calculate(tt.items)
			if result != tt.expected {
				t.Errorf("Calculate() = %v, want %v", result, tt.expected)
			}
		})
	}
}
