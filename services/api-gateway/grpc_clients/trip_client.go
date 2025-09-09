package grpc_clients

import (
	pb "CarpoolSharing/shared/proto/trip"
	"os"

	"google.golang.org/grpc"
)

type tripServiceClient struct {
	Client pb.TripServiceClient
}

func NewTripServiceClient() (*tripServiceClient, error) {
	tripServiceURL := os.Getenv("TRIP_SERVICE_URL")
	if tripServiceURL == "" {
		tripServiceURL = "trip-service:9093"
	}

	conn, err := grpc.NewClient(tripServiceURL)
	if err != nil {
		return nil, err
	}
}
