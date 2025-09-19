package api

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type TaskActionRequest struct {
	ID string `json:"id"`
}

type TaskResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"completed"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}