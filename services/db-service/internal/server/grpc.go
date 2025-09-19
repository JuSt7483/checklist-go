package server

import (
	pb "checklist-go/proto"
	"checklist-go/services/db-service/internal/storage"
	"context"
	"errors"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	pb.UnimplementedChecklistServiceServer
	storage *storage.Storage
}

func NewGRPCServer(storage *storage.Storage) *GRPCServer {
	return &GRPCServer{storage: storage}
}

func (s *GRPCServer) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.Task, error) {
	log.Printf("Received CreateTask request: title=%s", req.Title)

	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	task, err := s.storage.CreateTask(ctx, req.Title, req.Description)
	if err != nil {
		log.Printf("Error creating task: %v", err)
		return nil, status.Error(codes.Internal, "failed to create task")
	}

	log.Printf("Successfully created task with ID: %s", task.Id)
	return task, nil
}

func (s *GRPCServer) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	log.Println("Received ListTasks request")

	tasks, err := s.storage.ListTasks(ctx)
	if err != nil {
		log.Printf("Error listing tasks: %v", err)
		return nil, status.Error(codes.Internal, "failed to list tasks")
	}

	log.Printf("Successfully listed %d tasks", len(tasks))
	return &pb.ListTasksResponse{Tasks: tasks}, nil
}

func (s *GRPCServer) DeleteTask(ctx context.Context, req *pb.TaskActionRequest) (*pb.DeleteTaskResponse, error) {
	log.Printf("Received DeleteTask request for ID: %s", req.Id)

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "task ID is required")
	}

	err := s.storage.DeleteTask(ctx, req.Id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Printf("Task with ID %s not found for deletion", req.Id)
			return nil, status.Error(codes.NotFound, "task not found")
		}
		log.Printf("Error deleting task %s: %v", req.Id, err)
		return nil, status.Error(codes.Internal, "failed to delete task")
	}

	log.Printf("Successfully deleted task with ID: %s", req.Id)
	return &pb.DeleteTaskResponse{Success: true}, nil
}

func (s *GRPCServer) MarkTaskDone(ctx context.Context, req *pb.TaskActionRequest) (*pb.Task, error) {
	log.Printf("Received MarkTaskDone request for ID: %s", req.Id)

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "task ID is required")
	}

	err := s.storage.MarkTaskDone(ctx, req.Id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Printf("Task with ID %s not found for completion", req.Id)
			return nil, status.Error(codes.NotFound, "task not found")
		}
		log.Printf("Error marking task %s as done: %v", req.Id, err)
		return nil, status.Error(codes.Internal, "failed to mark task as done")
	}

	// Fetch the updated task to return it
	updatedTask, err := s.storage.GetTask(ctx, req.Id)
	if err != nil {
		// This should ideally not happen if the update succeeded
		log.Printf("Error fetching updated task %s: %v", req.Id, err)
		return nil, status.Error(codes.Internal, "failed to retrieve updated task")
	}

	log.Printf("Successfully marked task %s as done", req.Id)
	return updatedTask, nil
}