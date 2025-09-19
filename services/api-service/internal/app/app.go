package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	proto "checklist-go/proto"
	"checklist-go/services/api-service/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	httpServer *http.Server
	grpcServer *grpc.ClientConn
}

func New() (*App, error) {

	dbServiceAddr := "db-service:50051"

	conn, err := grpc.NewClient(dbServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db service: %w", err)
	}
	

	grpcClient := proto.NewChecklistServiceClient(conn)
	log.Println("Successfully connected to db-service")

	taskHandler := handlers.NewTaskHandler(grpcClient)


	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Post("/create", taskHandler.CreateTask)
	router.Get("/list", taskHandler.ListTasks)
	router.Delete("/delete", taskHandler.DeleteTask)
	router.Put("/done", taskHandler.MarkTaskDone)

	httpServerAddr := os.Getenv("HTTP_SERVER_ADDR")
	if httpServerAddr == "" {
		httpServerAddr = ":8080"
	}

	server := &http.Server{
		Addr:    httpServerAddr,
		Handler: router,
	}

	return &App{
		httpServer: server,
		grpcServer: conn,
	}, nil
}

func (a *App) Run() {
	go func(){
	log.Printf("HTTP server is listening on %s", a.httpServer.Addr)
		if err:= a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit 

	log.Println("Shutting down HTTP server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server failed to shutdown: %v", err)
	}

	if err := a.grpcServer.Close(); err != nil {
		log.Fatalf("gRPC client connection failed to close: %v", err)
	}

	log.Println("HTTP server and gRPC client connection closed")

}