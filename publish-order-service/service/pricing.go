package service

import (
	"math"

	"github.com/icl00ud/publish-order-service/pkg/model"
)

type PricingCalculator interface {
	Calculate(items []model.CartItem) int
}

type defaultPricing struct{}

func NewPricingCalculator() PricingCalculator {
	return defaultPricing{}
}

func (d defaultPricing) Calculate(items []model.CartItem) int {
	total := 0.0
	for _, it := range items {
		total += it.Price * float64(it.Quantity)
	}
	return int(math.Round(total))
}
