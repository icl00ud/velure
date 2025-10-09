package service

import "github.com/icl00ud/publish-order-service/internal/model"

type PricingCalculator interface {
	Calculate(items []model.CartItem) float64
}

type defaultPricing struct{}

func NewPricingCalculator() PricingCalculator {
	return defaultPricing{}
}

func (d defaultPricing) Calculate(items []model.CartItem) float64 {
	sum := 0.0
	for _, it := range items {
		sum += it.Price * float64(it.Quantity)
	}
	return sum
}
