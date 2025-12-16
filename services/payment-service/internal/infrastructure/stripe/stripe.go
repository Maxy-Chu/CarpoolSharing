package stripe

import (
	"CarpoolSharing/services/payment-service/internal/domain"
	"CarpoolSharing/services/payment-service/pkg/types"
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
)

type stripeClient struct {
	config *types.PaymentConfig
}

func NewStripeClient(cfg *types.PaymentConfig) domain.PaymentProcessor {
	// Initialize Stripe with the secret key - One time setup, that's why it's here
	stripe.Key = cfg.StripeSecretKey

	return &stripeClient{
		config: cfg,
	}
}

func (s *stripeClient) CreatePaymentSession(ctx context.Context, amount int64, currency string, metadata map[string]string) (string, error) {
	params := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(s.config.SuccessURL),
		CancelURL:  stripe.String(s.config.CancelURL),
		Metadata:   metadata, // use in the webhook to identify the payment
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Ride Payment"),
					},
					UnitAmount: stripe.Int64(amount),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
	}
	result, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create a payment session on stripe: %w", err)
	}

	return result.ID, nil
}
