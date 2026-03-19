package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "tz/gen/shipment/v1"
	"tz/internal/app"
	infrastructure "tz/internal/infrastructure/sqlite"
	grpchandler "tz/internal/transport/grpc"
)

func main() {
	port := getEnv("PORT", "50051")
	dsn := getEnv("DB_DSN", "./shipments.db")

	repo, err := infrastructure.NewSQLiteRepository(dsn)
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}

	svc := app.NewShipmentService(repo)

	handler := grpchandler.NewHandler(svc)

	grpcServer := grpc.NewServer()
	pb.RegisterShipmentServiceServer(grpcServer, handler)

	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	fmt.Printf("server started on :%s\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
