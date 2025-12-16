package types

import "time"

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCanceled  PaymentStatus = "canceled"
)

// Pyment represents a payment transaction
type Payment struct {
	ID              string        `json:"id"`
	TripID          string        `json:"trip_id"`
	UserID          string        `json:"user_id"`
	Amount          int64         `json:"amount"` // in cents
	Currency        string        `json:"currency"`
	Status          PaymentStatus `json:"status"`
	StripeSessionID string        `json:"stripe_session_id"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// PaymentIntent represents the intent to collect a payment
type PaymentIntent struct {
	ID              string    `json:"id"`
	TripID          string    `json:"trip_id"`
	UserID          string    `json:"user_id"`
	DriverID        string    `json:"driver_id"`
	Amount          int64     `json:"amount"`
	Currency        string    `json:"currency"`
	StripeSessionID string    `json:"stripe_session_id"`
	CreatedAt       time.Time `json:"created_at"`
}

// PaymentConfig holds the configuration for the payment service
type PaymentConfig struct {
	StripeSecretKey     string `json:"stripeSecretKey"`
	StripePulishableKey string `json:"stripePublishableKey"`
	StripeWebhookSecret string `json:"stripeWebhookSecret"`
	Currency            string `json:"currency"`
	SuccessURL          string `json:"successUrl"`
	CancelURL           string `json:"cancelUrl"`
}
