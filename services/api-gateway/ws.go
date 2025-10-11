package main

import (
	"CarpoolSharing/services/api-gateway/grpc_clients"
	"CarpoolSharing/shared/contracts"
	"CarpoolSharing/shared/proto/driver"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// upgrader to upgrade a httprequest into a websocket request
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Rider handler
func handleRidersWebSocket(w http.ResponseWriter, r *http.Request) {
	// upgrader send httprequest and httpresponsewriter and receive a WEBSOCKET CONNECTION pointer
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Rider WebSocket upgrade failed: %v", err)
		return
	}

	defer conn.Close()

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Println("No user ID provided")
	}

	// infinite loops for listening
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
func handleDriversWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
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

	ctx := r.Context()

	// Call Driver Service
	driverService, err := grpc_clients.NewdriverServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	// Closing connections
	defer func() {
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
		log.Fatal(err)
	}

	msg := contracts.WSMessage{
		Type: "driver.cmd.register",
		Data: driverData.Driver,
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}

}
