package ports

import (
	"context"
	"time"

	"todoapp/services/task-service/internal/domain/entities"
)

type TaskFilter struct {
	Statuses   []entities.TaskStatus
	Priorities []entities.TaskPriority
	CategoryID *int64
	Search     string
	DueFrom    *time.Time
	DueTo      *time.Time
	Limit      int
	Offset     int
}

type TaskRepository interface {
	CreateTask(ctx context.Context, task *entities.Task) error
	UpdateTask(ctx context.Context, task *entities.Task) error
	SoftDeleteTask(ctx context.Context, userID, taskID int64, deletedAt time.Time) error
	GetTask(ctx context.Context, userID, taskID int64) (*entities.Task, error)
	ListTasks(ctx context.Context, userID int64, filter TaskFilter) ([]entities.Task, error)

	CreateCategory(ctx context.Context, category *entities.Category) error
	ListCategories(ctx context.Context, userID int64) ([]entities.Category, error)
	GetCategory(ctx context.Context, userID, categoryID int64) (*entities.Category, error)
	DeleteCategory(ctx context.Context, userID, categoryID int64) error

	CreateComment(ctx context.Context, comment *entities.TaskComment) error
	ListComments(ctx context.Context, userID, taskID int64) ([]entities.TaskComment, error)
}
