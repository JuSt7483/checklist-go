package server

import (
	"context"
	"errors"

	pb "checklist-go/proto"
	"checklist-go/services/db-service/internal/storage"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ChecklistServer struct {
	pb.UnimplementedChecklistServiceServer
	storage *storage.Storage
}

func NewChecklistServer(storage *storage.Storage) *ChecklistServer {
	return &ChecklistServer{
		storage: storage,
	}
}

// --- CreateTask ---
func (c *ChecklistServer) CreateTask( ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	task, err := c.storage.CreateTask(ctx, req.GetTitle(), req.GetDescription())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create task: %v", err)
	}
	return &pb.CreateTaskResponse{Task: task}, nil
}

func (c *ChecklistServer) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	tasks, err := c.storage.ListTasks(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list tasks: %v", err)
	}

	return &pb.ListTasksResponse{Tasks: tasks}, nil
}

func (c *ChecklistServer) CompleteTask(ctx context.Context, req *pb.CompleteTaskRequest) (*emptypb.Empty, error) {
	err := c.storage.CompleteTask(ctx, req.Id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "task with id %s not found", req.GetId())
		}
		return nil, status.Errorf(codes.Internal, "failed to complete task: %v", err)
	}
	return &emptypb.Empty{}, nil

}

func (c *ChecklistServer) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*emptypb.Empty, error) {
	err := c.storage.DeleteTask(ctx, req.Id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, "task with id %s not found", req.GetId())
		}
		return nil, status.Errorf(codes.Internal, "failed to delete task: %v", err)
	}
	
	return &emptypb.Empty{}, nil
}