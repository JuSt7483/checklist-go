package storage

import (
	pb "checklist-go/proto"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"
)


type Storage struct {
	db *pgxpool.Pool
}

func NewStorage(dsn string) (*Storage, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}
	return &Storage{db: pool}, nil
}

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) CreateTask(ctx context.Context, title string, description string) (*pb.Task, error) {
	id := uuid.New()

	query := `INSERT INTO tasks (id, title, description) VALUES ($1, $2, $3) RETURNING created_at, updated_at`

	var createdAt, updatedAt time.Time

	err := s.db.QueryRow(ctx, query, id, title, description).Scan(&createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return &pb.Task{
		Id:          id.String(),
		Title:       title,
		Description: description,
		Done:   false,
		CreatedAt:   timestamppb.New(createdAt),
		UpdatedAt:   timestamppb.New(updatedAt),
	}, nil
}

func (s *Storage) ListTasks(ctx context.Context) ([]*pb.Task, error) {
	query := `SELECT id, title, description, done, created_at, updated_at FROM tasks ORDER BY created_at desc`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*pb.Task
	for rows.Next() {
		var id uuid.UUID
		var title, description string
		var done bool
		var createdAt, updatedAt time.Time

		if err = rows.Scan(&id, &title, &description, &done, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		tasks = append(tasks, &pb.Task{
			Id:          id.String(),
			Title:       title,
			Description: description,
			Done:   done,
			CreatedAt:   timestamppb.New(createdAt),
			UpdatedAt:   timestamppb.New(updatedAt),
		})
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over tasks: %w", err)
	}
	return tasks, nil
} 

func (s *Storage) GetTask(ctx context.Context, id string) (*pb.Task, error) {
	query := `SELECT id, title, description, done, created_at, updated_at FROM tasks WHERE id = $1`

	var task pb.Task
	var uid uuid.UUID
	var created_at, updated_at time.Time

	err := s.db.QueryRow(ctx, query, id).Scan(&uid, &task.Title, &task.Description, &task.Done, &created_at, &updated_at)
	if err != nil {
		if err.Error() == "no rows in result set"{
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	task.Id = uid.String()
	task.CreatedAt = timestamppb.New(created_at)
	task.UpdatedAt = timestamppb.New(updated_at)

	return &task, nil
}

func (s *Storage) MarkTaskDone(ctx context.Context, id string) error {
	query := `UPDATE tasks SET done = true, updated_at = NOW() WHERE id = $1`
	cmdTag, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark task done: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *Storage) CompleteTask (ctx context.Context, id string) error {
	query := `UPDATE tasks SET completed = true, updated_at = NOW() WHERE id = $1`
	cmdTag, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *Storage) DeleteTask (ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = $1`
	cmdTag, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}