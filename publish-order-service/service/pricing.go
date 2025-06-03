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
	sum := 0
	for _, it := range items {
		sum += int(math.Round(it.Price * float64(it.Quantity)))
	}
	return sum
}
