package payment

import (
	"context"
	"errors"
	"fmt"

	stripe "github.com/stripe/stripe-go/v82"
)

// StripeProcessor charges orders through the Stripe API (test mode with an
// sk_test_ key). The order ID doubles as the Stripe idempotency key, so a
// redelivered message can never double-charge.
type StripeProcessor struct {
	client *stripe.Client
}

type StripeOption func(*stripeConfig)

type stripeConfig struct {
	baseURL string
}

// WithBaseURL points the client at a fake Stripe server (tests only).
func WithBaseURL(url string) StripeOption {
	return func(c *stripeConfig) { c.baseURL = url }
}

func NewStripeProcessor(apiKey string, opts ...StripeOption) *StripeProcessor {
	var cfg stripeConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	var clientOpts []stripe.ClientOption
	if cfg.baseURL != "" {
		backends := stripe.NewBackends(nil)
		backends.API = stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
			URL: stripe.String(cfg.baseURL),
		})
		clientOpts = append(clientOpts, stripe.WithBackends(backends))
	}

	return &StripeProcessor{client: stripe.NewClient(apiKey, clientOpts...)}
}

func (p *StripeProcessor) Charge(ctx context.Context, orderID string, amountCents int64) error {
	params := &stripe.PaymentIntentCreateParams{
		Params: stripe.Params{
			IdempotencyKey: stripe.String(orderID),
		},
		Amount:   stripe.Int64(amountCents),
		Currency: stripe.String(string(stripe.CurrencyBRL)),
		// Test-mode payment method that always succeeds; declines are
		// exercised with Stripe's special test cards.
		PaymentMethod: stripe.String("pm_card_visa"),
		Confirm:       stripe.Bool(true),
		AutomaticPaymentMethods: &stripe.PaymentIntentCreateAutomaticPaymentMethodsParams{
			Enabled:        stripe.Bool(true),
			AllowRedirects: stripe.String("never"),
		},
	}

	intent, err := p.client.V1PaymentIntents.Create(ctx, params)
	if err != nil {
		var stripeErr *stripe.Error
		if errors.As(err, &stripeErr) && stripeErr.Type == stripe.ErrorTypeCard {
			return &PermanentError{Reason: stripeErr.Msg}
		}
		return fmt.Errorf("stripe charge: %w", err)
	}

	if intent.Status != stripe.PaymentIntentStatusSucceeded {
		return &PermanentError{Reason: fmt.Sprintf("payment intent status %s", intent.Status)}
	}
	return nil
}
