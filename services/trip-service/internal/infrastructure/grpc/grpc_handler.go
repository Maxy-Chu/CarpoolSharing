package grpc

import (
	"CarpoolSharing/services/trip-service/internal/domain"
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
	Service domain.TripService
}

func NewGRPCHandler(server *grpc.Server, service domain.TripService) *gRPCHandler {
	handler := &gRPCHandler{
		Service: service,
	}

	pb.RegisterTripServiceServer(server, handler)
	return handler
}

func (h *gRPCHandler) PreviewTrip(ctx context.Context, req *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {
	pickupCoord := &types.Coordinate{
		Latitude:  req.StartLocation.GetLatitude(),
		Longitude: req.StartLocation.GetLongitude(),
	}
	destinationCoord := &types.Coordinate{
		Latitude:  req.EndLocation.GetLatitude(),
		Longitude: req.EndLocation.GetLongitude(),
	}

	route, err := h.Service.GetRoute(ctx, pickupCoord, destinationCoord)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to get route: %v", err)
	}

	return &pb.PreviewTripResponse{
		Route:     route.ToProto(),
		RideFares: []*pb.RideFare{},
	}, nil
}
