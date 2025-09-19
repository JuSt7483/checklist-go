package handlers

import (
	proto "checklist-go/proto"
	"checklist-go/services/api-service/internal/api"
	"context"
	"encoding/json"
	"log" // Добавляем импорт пакета log
	"net/http"
	"time"

	"google.golang.org/grpc/status"
)


type TaskHandler struct {
	grpcClient proto.ChecklistServiceClient
}

func NewTaskHandler(grpcClient proto.ChecklistServiceClient) *TaskHandler {
	return &TaskHandler{
		grpcClient: grpcClient,
	}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req api.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	grpcReq := &proto.CreateTaskRequest{
		Title:       req.Title,
		Description: req.Description,
	}

	grpcRes, err := h.grpcClient.CreateTask(ctx, grpcReq)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			// Если это не gRPC статус-ошибка, логируем сырую ошибку
			log.Printf("Failed to create task: non-gRPC error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		// Логируем код и сообщение gRPC ошибки
		log.Printf("Failed to create task: gRPC error code=%s, message=%s", st.Code().String(), st.Message())
		http.Error(w, st.Message(), http.StatusInternalServerError)
		return
	}

	res := &api.TaskResponse{
		ID:          grpcRes.Id,
		Title:       grpcRes.Title,
		Description: grpcRes.Description,
		Done:        grpcRes.Done,
		CreatedAt:   grpcRes.CreatedAt.AsTime().Format(time.RFC3339),
		UpdatedAt:   grpcRes.UpdatedAt.AsTime().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	grpcRes, err := h.grpcClient.ListTasks(ctx, &proto.ListTasksRequest{})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	var tasks []*api.TaskResponse
	for _, task := range grpcRes.Tasks {
		tasks = append(tasks, &api.TaskResponse{
			ID:          task.Id,
			Title:       task.Title,
			Description: task.Description,
			Done:        task.Done,
			CreatedAt:   task.CreatedAt.AsTime().Format(time.RFC3339),
			UpdatedAt:   task.UpdatedAt.AsTime().Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
} 

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	var req api.TaskActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	_, err := h.grpcClient.DeleteTask(ctx, &proto.TaskActionRequest{Id: req.ID})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}



func (h *TaskHandler) MarkTaskDone(w http.ResponseWriter, r *http.Request) {
	var req api.TaskActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	_, err := h.grpcClient.MarkTaskDone(ctx, &proto.TaskActionRequest{Id: req.ID})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


func handleGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	http.Error(w, st.Message(), http.StatusInternalServerError)
}