package main

import (
	"CarpoolSharing/shared/messaging"
	"context"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// struct, constructor, listen

type tripConsumer struct {
	rabbitmq *messaging.RabbitMQ
}

func NewTripConsumer(rabbitmq *messaging.RabbitMQ) *tripConsumer {
	return &tripConsumer{
		rabbitmq: rabbitmq,
	}
}

// method for driver-service/main.go, create a goroutine to listen the message on the background
func (c *tripConsumer) Listen() error {
	//multiple consumer
	return c.rabbitmq.ConsumeMessages(messaging.FindAvailableDriversQueue, func(ctx context.Context, msg amqp.Delivery) error {
		log.Printf("driver received message: %v", msg)
		return nil
	})
}
