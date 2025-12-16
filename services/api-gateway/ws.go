package main

import (
	"CarpoolSharing/services/api-gateway/grpc_clients"
	"CarpoolSharing/shared/contracts"
	"CarpoolSharing/shared/messaging"
	"CarpoolSharing/shared/proto/driver"
	"encoding/json"
	"log"
	"net/http"
)

// upgrader to upgrade a httprequest into a websocket request
var (
	connManager = messaging.NewConnectionManager()
)

// Rider handler
func handleRidersWebSocket(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
	// upgrader send httprequest and httpresponsewriter and receive a WEBSOCKET CONNECTION pointer
	conn, err := connManager.Upgrade(w, r)
	if err != nil {
		log.Printf("Rider WebSocket upgrade failed: %v", err)
		return
	}

	defer conn.Close()

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Println("No user ID provided")
	}

	/* Add Connection To Manager */
	connManager.Add(userID, conn)
	defer connManager.Remove(userID)

	/* Initialize A Queue Consumer */
	queues := []string{
		messaging.NotifyDriversNoDriverFoundQueue,
		messaging.NotifyDriversNoDriverFoundQueue,
		messaging.NotifyPaymentEventSessionCreatedQueue,
	}

	for _, q := range queues {
		consumer := messaging.NewQueueConsumer(rb, connManager, q)

		if err := consumer.Start(); err != nil {
			log.Printf("Failed to start consumer for queue: %s; err: %v", q, err)
		}
	}

	/* Infinite Loops For Listening */
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		log.Printf("Rider received message: %s", message)
	}
}

// Driver handler
func handleDriversWebSocket(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
	conn, err := connManager.Upgrade(w, r)
	if err != nil {
		log.Printf("Driver WebSocket upgrade failed: %v", err)
		return
	}

	defer conn.Close()

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Println("No user ID provided")
		return
	}

	packageSlug := r.URL.Query().Get("packageSlug")
	if packageSlug == "" {
		log.Println("No packageSlug provided")
		return
	}

	// Add connection to manager
	connManager.Add(userID, conn)

	ctx := r.Context()

	// Call Driver Service
	driverService, err := grpc_clients.NewdriverServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	// Closing connections
	defer func() {
		connManager.Remove(userID)

		driverService.Client.UnregisterDriver(ctx, &driver.RegisterDriverRequest{
			DriverID:    userID,
			PackageSlug: packageSlug,
		})
		driverService.Close()
		log.Println("Driver unregistered: ", userID)
	}()

	driverData, err := driverService.Client.RegisterDriver(ctx, &driver.RegisterDriverRequest{
		DriverID:    userID,
		PackageSlug: packageSlug,
	})
	if err != nil {
		log.Printf("Error registering driver: %v", err)
		return
	}

	if err := connManager.SendMessage(userID, contracts.WSMessage{
		Type: contracts.DriverCmdRegister,
		Data: driverData.Driver,
	}); err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}

	/* Initialize A Queue Consumer */
	queues := []string{
		messaging.DriverCmdTripRequestQueue,
	}

	for _, q := range queues {
		consumer := messaging.NewQueueConsumer(rb, connManager, q)

		if err := consumer.Start(); err != nil {
			log.Printf("Failed to start consumer for queue: %s; err: %v", q, err)
		}
	}

	/* Read Message + Publish Message */
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		// save the data from unmarshaled message
		type driverMessage struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}

		var driverMsg driverMessage
		if err := json.Unmarshal(message, &driverMsg); err != nil {
			log.Printf("Error unmarshaling driver message (api-gateway/ws.go/HandleDriversWebsocket): %v", err)
		}

		// Handle the different message type
		switch driverMsg.Type {
		case contracts.DriverCmdLocation:
			// Handle driver location update in the future
			continue
		case contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline:
			// Forward the message to RabbitMQ
			if err := rb.PublishMessage(ctx, driverMsg.Type, contracts.AmqpMessage{
				OwnerID: userID,
				Data:    driverMsg.Data,
			}); err != nil {
				log.Printf("Error publishing message to RabbitMQ (api-gateway/ws.go/HandleDriversWebsocket): %v", err)
			}
		default:
			log.Printf("Unknown message type (api-gateway/ws.go/HandleDriversWebsocket): %s", driverMsg.Type)
		}

	}

}
