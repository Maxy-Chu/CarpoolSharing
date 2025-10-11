package events

import (
	"CarpoolSharing/shared/messaging"
	"context"
)

type TripEventPublisher struct {
	rabbitmq *messaging.RabbitMQ
}

func NewTripEventPublisher(rabbitmq *messaging.RabbitMQ) *TripEventPublisher {
	return &TripEventPublisher{
		rabbitmq: rabbitmq,
	}
}

func (p *TripEventPublisher) PublishTripCreated(ctx context.Context) error {
	return p.rabbitmq.PublishMessage(ctx, "hello", "hello world")
}
