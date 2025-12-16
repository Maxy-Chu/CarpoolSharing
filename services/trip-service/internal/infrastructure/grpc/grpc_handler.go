package grpc

import (
	"CarpoolSharing/services/trip-service/internal/domain"
	"CarpoolSharing/services/trip-service/internal/infrastructure/events"
	pb "CarpoolSharing/shared/proto/trip"
	"CarpoolSharing/shared/types"
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pb.UnimplementedTripServiceServer
	Service   domain.TripService
	publisher *events.TripEventPublisher
}

func NewGRPCHandler(server *grpc.Server, service domain.TripService, publisher *events.TripEventPublisher) *gRPCHandler {
	handler := &gRPCHandler{
		Service:   service,
		publisher: publisher,
	}

	pb.RegisterTripServiceServer(server, handler)
	return handler
}

func (h *gRPCHandler) PreviewTrip(ctx context.Context, req *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {
	// gRPC Connect with pb and get the data
	pickupCoord := &types.Coordinate{
		Latitude:  req.StartLocation.GetLatitude(),
		Longitude: req.StartLocation.GetLongitude(),
	}
	destinationCoord := &types.Coordinate{
		Latitude:  req.EndLocation.GetLatitude(),
		Longitude: req.EndLocation.GetLongitude(),
	}

	userID := req.GetUserID()

	// gRPC Connect with domain logic and parse the data
	route, err := h.Service.GetRoute(ctx, pickupCoord, destinationCoord)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to get route: %v", err)
	}

	// Estimate the ride based on the route
	estimatedFares := h.Service.EstimatePackagesPriceWithRoute(route)
	// Store the data to DB
	fares, err := h.Service.GenerateTripFares(ctx, estimatedFares, userID, route)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate the ride fares: %v", err)
	}

	// gRPC Connect with pb and return the data
	return &pb.PreviewTripResponse{
		Route:     route.ToProto(),
		RideFares: domain.ToRideFaresProto(fares),
	}, nil
}

func (h *gRPCHandler) CreateTrip(ctx context.Context, req *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {
	fareID := req.GetRideFareID()
	userID := req.GetUserID()
	// Fetch and validate the fare
	rideFare, err := h.Service.GetAndValidateFare(ctx, fareID, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate the fare: %v", err)
	}

	// Call create trip
	trip, err := h.Service.CreateTrip(ctx, rideFare)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create the trip: %v", err)
	}
	// TODO We also need to initialize
	// Add a comment at the end of the function to publish an event on the Async Comms module
	if err := h.publisher.PublishTripCreated(ctx, trip); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish the trip created event: %v", err)
	}

	return &pb.CreateTripResponse{
		TripID: trip.ID.Hex(),
	}, nil

}
