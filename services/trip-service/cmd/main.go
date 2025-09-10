package main

import (
	"CarpoolSharing/services/trip-service/internal/infrastructure/grpc"
	"CarpoolSharing/services/trip-service/internal/infrastructure/repository"
	"CarpoolSharing/services/trip-service/internal/service"
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9093"

func main() {
	inmemRepo := repository.NewInmemRepository()
	svc := service.NewService(inmemRepo)

	// create context and cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	// listen to TCP endpoint
	lis, err := net.Listen("tcp", GrpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Starting gRPC Server, initialize grpc handler
	grpcServer := grpcserver.NewServer()
	grpc.NewGRPCHandler(grpcServer, svc)

	log.Printf("Starting gRPC Trip Service on port %s", lis.Addr().String())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("failed to serve: %v", err)
			cancel()
		}
	}()

	// wait for the shutdown signal
	<-ctx.Done()
	log.Println("Shutting down the server...")
	grpcServer.GracefulStop()
}
