package events

import (
	"CarpoolSharing/services/trip-service/internal/domain"
	"CarpoolSharing/shared/contracts"
	"CarpoolSharing/shared/messaging"
	"context"
	"encoding/json"
	"fmt"
	"log"

	pbd "CarpoolSharing/shared/proto/driver"

	"github.com/rabbitmq/amqp091-go"
)

type driverConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  domain.TripService
}

func NewDriverConsumer(rabbitmq *messaging.RabbitMQ, service domain.TripService) *driverConsumer {
	return &driverConsumer{
		rabbitmq: rabbitmq,
		service:  service,
	}
}

// method for driver-service/main.go, create a goroutine to listen the message on the background
func (c *driverConsumer) Listen() error {
	//multiple consumer
	return c.rabbitmq.ConsumeMessages(messaging.DriverTripResponseQueue, func(ctx context.Context, msg amqp091.Delivery) error {
		var driverEvent contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &driverEvent); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			return err
		}

		var payload messaging.DriverTripResponseData
		if err := json.Unmarshal(driverEvent.Data, &payload); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			return err
		}

		log.Printf("driver response received message: %+v", payload)

		// find messages
		switch msg.RoutingKey {
		case contracts.DriverCmdTripAccept:
			if err := c.handleTripAccepted(ctx, payload.TripID, payload.Driver); err != nil {
				log.Printf("Failed to handle the trip accept: %v", err)
				return err
			}

		case contracts.DriverCmdTripDecline:
			if err := c.handleTripDeclined(ctx, payload.TripID, payload.RiderID); err != nil {
				log.Printf("Failed to handle the trip decline: %v", err)
				return err
			}
			return nil
		}

		log.Printf("unknown trip event %+v", payload) // if switch doesn't find any trip

		return nil
	})
}

func (c *driverConsumer) handleTripAccepted(ctx context.Context, tripID string, driver *pbd.Driver) error {
	// 1. Fetch the trip - validate if the actual trip exists
	trip, err := c.service.GetTripByID(ctx, tripID)
	if err != nil {
		log.Printf("Failed to get the trip (driver_consumer): %v", err)
		return err
	}
	if trip == nil {
		return fmt.Errorf("trip was not found, wrong tripID: %s", tripID)
	}

	// 2. Update the trip
	if err := c.service.UpdateTrip(ctx, tripID, "Accepted", driver); err != nil {
		log.Printf("Failed to update the trip (driver_consumer): %v", err)
		return err
	}

	trip, err = c.service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}
	// 3. Marshal the trip data
	marshalledTrip, err := json.Marshal(trip)
	if err != nil {
		return err
	}

	// 4. Notify the Rider that Driver has been assigned -> publish the event to RabbitMQ
	if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventDriverAssigned, contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    marshalledTrip,
	}); err != nil {
		return err
	}

	// 5. Notify the payment service to start a payment link
	marshalledPayload, err := json.Marshal(messaging.PaymentTripResponseData{
		TripID:   tripID,
		UserID:   trip.UserID,
		DriverID: driver.Id,
		Amount:   trip.RideFare.TotalPriceInCents,
		Currency: "USD",
	})

	if err != nil {
		log.Printf("Failed to marshal the Payment Trip Response Data(driver_consumer): %v", err)
		return nil
	}

	if err := c.rabbitmq.PublishMessage(ctx, contracts.PaymentCmdCreateSession, contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    marshalledPayload,
	}); err != nil {
		log.Printf("Failed to publish payment message to payment service via rabbitmq(driver_consumer): %v", err)
		return nil
	}
	return nil
}

func (c *driverConsumer) handleTripDeclined(ctx context.Context, tripID, riderID string) error {
	// 1. Fetch the trip - validate if the actual trip exists
	trip, err := c.service.GetTripByID(ctx, tripID)
	if err != nil {
		log.Printf("Failed to get the trip (driver_consumer): %v", err)
		return err
	}

	// 2. Create new payload and Marshal it
	newPayload := messaging.TripEventData{
		Trip: trip.ToProto(),
	}

	marshalledPayload, err := json.Marshal(newPayload)
	if err != nil {
		return err
	}

	// 3. Driver not interested -> Publish the event to RabbitMQ
	if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventDriverNotInterested, contracts.AmqpMessage{
		OwnerID: riderID,
		Data:    marshalledPayload,
	}); err != nil {
		return err
	}

	return nil
}
