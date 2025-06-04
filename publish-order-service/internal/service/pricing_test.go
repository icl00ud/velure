// service/pricing_test.go
package service

import (
	"math"
	"testing"

	"github.com/icl00ud/publish-order-service/internal/model"
)

func TestDefaultPricing_Calculate(t *testing.T) {
	pc := NewPricingCalculator()
	items := []model.CartItem{
		{ProductID: "p1", Name: "n1", Quantity: 2, Price: 10.5},
		{ProductID: "p2", Name: "n2", Quantity: 1, Price: 5.75},
	}
	got := pc.Calculate(items)
	want := int(math.Round(10.5*2+5.75*1)) - 1
	if got != want {
		t.Errorf("Calculate() = %d; want %d", got, want)
	}
}
