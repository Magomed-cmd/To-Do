package ports

import (
	"context"
	"time"

	"todoapp/services/task-service/internal/domain/entities"
)

type CreateTaskInput struct {
	UserID      int64
	Title       string
	Description string
	Status      entities.TaskStatus
	Priority    entities.TaskPriority
	DueDate     *time.Time
	CategoryID  *int64
}

type UpdateTaskInput struct {
	UserID        int64
	TaskID        int64
	Title         *string
	Description   *string
	Status        *entities.TaskStatus
	Priority      *entities.TaskPriority
	DueDate       *time.Time
	ClearDueDate  bool
	CategoryID    *int64
	ClearCategory bool
}

type AddCommentInput struct {
	UserID  int64
	TaskID  int64
	Content string
}

type CreateCategoryInput struct {
	UserID int64
	Name   string
}

type TaskService interface {
	CreateTask(ctx context.Context, input CreateTaskInput) (*entities.Task, error)
	UpdateTask(ctx context.Context, input UpdateTaskInput) (*entities.Task, error)
	UpdateTaskStatus(ctx context.Context, userID, taskID int64, status entities.TaskStatus) (*entities.Task, error)
	DeleteTask(ctx context.Context, userID, taskID int64) error
	GetTask(ctx context.Context, userID, taskID int64) (*entities.Task, error)
	ListTasks(ctx context.Context, userID int64, filter TaskFilter) ([]entities.Task, error)

	// ExportTasks exports all user tasks in the specified format.
	// Returns the file content, filename, and any error.
	ExportTasks(ctx context.Context, userID int64, format entities.ExportFormat) ([]byte, string, error)

	CreateCategory(ctx context.Context, input CreateCategoryInput) (*entities.Category, error)
	ListCategories(ctx context.Context, userID int64) ([]entities.Category, error)
	DeleteCategory(ctx context.Context, userID, categoryID int64) error

	AddComment(ctx context.Context, input AddCommentInput) (*entities.TaskComment, error)
	ListComments(ctx context.Context, userID, taskID int64) ([]entities.TaskComment, error)
}
