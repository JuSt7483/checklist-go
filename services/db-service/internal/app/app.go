package app

import (
	"checklist-go/services/db-service/internal/server"
	"checklist-go/services/db-service/internal/storage"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "checklist-go/proto"
)

type App struct {
	grpcServer *grpc.Server
	storage *storage.Storage
	lis net.Listener
}

func New() (*App, error) {
	st, err := initStorage()
	if err != nil {
		return nil, err
	}

	lis, grpcServer, err := initGRPCServer(st)
	if err != nil {
		return nil, err
	}

	return &App{
		grpcServer: grpcServer,
		storage: st,
		lis: lis,
	}, nil
}

func initStorage() (*storage.Storage, error) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://checklist_user:checklist_pass@localhost:5432/checklist_db"
	}
	
	st, err := storage.NewStorage(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}
	log.Println("Successfully connected to the database")
	return st, nil
}

func initGRPCServer(st *storage.Storage) (net.Listener, *grpc.Server, error) {
	grpcAddr := os.Getenv("DB_GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = ":50051"
	}

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen: %w", err)
	}

	log.Println("Successfully created gRPC server on", grpcAddr)

	grpcServer := grpc.NewServer()
	checkListServer := server.NewChecklistServer(st)
	pb.RegisterChecklistServiceServer(grpcServer, checkListServer)

	reflection.Register(grpcServer)

	return lis, grpcServer, nil
}

func (a *App) Run() {
	go func (){
		log.Println("db-service gRPC server started on", a.lis.Addr().String())
		if err := a.grpcServer.Serve(a.lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down db-service gRPC server...")

	a.grpcServer.GracefulStop()
	log.Println("db-service gRPC server stopped")

	a.storage.Close()
	log.Println("Storage closed")
}