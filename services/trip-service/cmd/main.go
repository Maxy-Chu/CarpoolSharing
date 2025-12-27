package main

import (
	"CarpoolSharing/services/trip-service/internal/infrastructure/events"
	"CarpoolSharing/services/trip-service/internal/infrastructure/grpc"
	"CarpoolSharing/services/trip-service/internal/infrastructure/repository"
	"CarpoolSharing/services/trip-service/internal/service"
	"CarpoolSharing/shared/env"
	"CarpoolSharing/shared/messaging"
	"CarpoolSharing/shared/tracing"
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

// Service Address
var GrpcAddr = ":9093"

func main() {
	// TODO: Replace Immemory Repository with Database
	inmemRepo := repository.NewInmemRepository()
	svc := service.NewService(inmemRepo)

	// Create context and cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Tracing
	tracerCfg := tracing.Config{
		ServiceName:    "trip-service",
		Environment:    env.GetString("ENVIRONMENT", "development"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	}
	sh, err := tracing.InitTracer(tracerCfg)
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer sh(ctx)

	// Listen for Ctrl+C graceful shutdown signal
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

	// RabbitMQ connection
	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()
	log.Println("Starting RabbitMQ connection")

	// Starting trip event publisher
	publisher := events.NewTripEventPublisher(rabbitmq)

	// Starting driver consumer
	driverConsumer := events.NewDriverConsumer(rabbitmq, svc)
	go driverConsumer.Listen()

	// Starting payment consumer
	paymentConsumer := events.NewPaymentConsumer(rabbitmq, svc)
	go paymentConsumer.Listen()

	// Starting gRPC Server, initialize grpc handler
	grpcServer := grpcserver.NewServer(tracing.WithTracingInterceptors()...)
	grpc.NewGRPCHandler(grpcServer, svc, publisher)
	log.Printf("Starting gRPC Trip Service on port %s", lis.Addr().String())
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("failed to serve with gRPC: %v", err)
			cancel()
		}
	}()

	// Waiting for shutdown signal from background
	<-ctx.Done()
	log.Println("Shutting down the server...")
	grpcServer.GracefulStop()
}
