package app

import (
	"checklist-go/services/db-service/internal/server"
	"checklist-go/services/db-service/internal/storage"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	pb "checklist-go/proto"
)

type App struct {
	grpcServer *grpc.Server
	storage *storage.Storage
}

func New() (*App, error) {
	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		return nil, fmt.Errorf("DB_DSN environment variable is not set")
	}

	st, err := storage.NewStorage(dbDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	grpcSrv := grpc.NewServer()
	checkListServer := server.NewGRPCServer(st)
	pb.RegisterChecklistServiceServer(grpcSrv, checkListServer)

	return &App{
		grpcServer: grpcSrv,
		storage: st,
	}, nil
}

func (a *App) Run() {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("gRPC server is listening on port %s", port)
	if err := a.grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}