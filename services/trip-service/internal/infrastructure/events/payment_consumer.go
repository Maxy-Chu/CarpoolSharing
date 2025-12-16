package events

import (
	"CarpoolSharing/services/trip-service/internal/domain"
	"CarpoolSharing/shared/contracts"
	"CarpoolSharing/shared/messaging"
	"context"
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

type paymentConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  domain.TripService
}

func NewPaymentConsumer(rb *messaging.RabbitMQ, svc domain.TripService) *paymentConsumer {
	return &paymentConsumer{
		rabbitmq: rb,
		service:  svc,
	}
}

func (c *paymentConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(messaging.NotifyPaymentSuccessQueue, func(ctx context.Context, msg amqp091.Delivery) error {
		var message contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &message); err != nil {
			log.Printf("failed to unmarshal message: %v", err)
			return err
		}

		var payload messaging.PaymentStatusUpdateData
		if err := json.Unmarshal(message.Data, &payload); err != nil {
			log.Printf("failed to unmarshal payment status update data: %v", err)
			return err
		}

		log.Printf("Trip has been completed and paid.")

		return c.service.UpdateTrip(
			ctx,
			payload.TripID,
			"Paid",
			nil,
		)
	})

}
