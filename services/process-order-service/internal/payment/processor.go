// Package payment abstracts the payment step of order processing.
// Two implementations exist: StripeProcessor (Stripe test mode) and
// SimulatedProcessor (latency-only fake, used when no STRIPE_API_KEY is set
// so local development needs no Stripe account).
package payment

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"
)

// Processor charges a payment for an order. Implementations must be
// idempotent per orderID: retrying a Charge for the same order must not
// charge twice.
type Processor interface {
	Charge(ctx context.Context, orderID string, amountCents int64) error
}

// PermanentError marks a payment failure that retrying cannot fix
// (e.g. card declined). The caller should fail the order instead of retrying.
type PermanentError struct {
	Reason string
}

func (e *PermanentError) Error() string {
	return "payment permanently failed: " + e.Reason
}

func asPermanent(err error, target **PermanentError) bool {
	return errors.As(err, target)
}

// SimulatedProcessor mimics payment latency and always succeeds.
type SimulatedProcessor struct {
	maxLatency time.Duration
}

// NewSimulatedProcessor creates a fake processor. maxLatency bounds the
// random artificial delay; 0 disables it (useful in tests).
func NewSimulatedProcessor(maxLatency time.Duration) *SimulatedProcessor {
	return &SimulatedProcessor{maxLatency: maxLatency}
}

func (s *SimulatedProcessor) Charge(ctx context.Context, orderID string, amountCents int64) error {
	if s.maxLatency <= 0 {
		return nil
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(s.maxLatency)))
	if err != nil {
		return fmt.Errorf("generate random latency: %w", err)
	}
	select {
	case <-time.After(time.Duration(n.Int64())):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
