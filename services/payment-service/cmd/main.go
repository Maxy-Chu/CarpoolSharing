package main

import (
	"CarpoolSharing/services/payment-service/internal/events"
	"CarpoolSharing/services/payment-service/internal/infrastructure/stripe"
	"CarpoolSharing/services/payment-service/internal/service"
	"CarpoolSharing/services/payment-service/pkg/types"
	"CarpoolSharing/shared/env"
	"CarpoolSharing/shared/messaging"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var GrpcAddr = env.GetString("GRPC_ADDR", ":9004")

func main() {
	rabbitmqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	appURL := env.GetString("APP_URL", "http://localhost:3000")

	// Stripe config
	stripeCfg := &types.PaymentConfig{
		StripeSecretKey: env.GetString("STRIPE_SECRET_KEY", ""),
		SuccessURL:      env.GetString("STRIPE_SUCCESS_URL", appURL+"?payment=success"),
		CancelURL:       env.GetString("STRIPE_CANCEL_URL", appURL+"?payment=failed"),
	}

	if stripeCfg.StripeSecretKey == "" {
		log.Fatalf("STRIPE_SECRET_KEY is not set")
		return
	}

	// Stripe Processor
	paymentProcessor := stripe.NewStripeClient(stripeCfg)

	// Payment Service
	svc := service.NewPaymentService(paymentProcessor)

	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitmqURI)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitmq.Close()

	log.Println("Starting RabbitMQ connection...")

	// Trip Consumer
	tripConsumer := events.NewTripConsumer(rabbitmq, svc)
	go tripConsumer.Listen()

	// Wait for sutdown signal
	<-ctx.Done()
	log.Println("Shutting down payment service...")
}
