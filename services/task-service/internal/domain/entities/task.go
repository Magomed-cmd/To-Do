package entities

import "time"

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusArchived   TaskStatus = "archived"
)

type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

type Task struct {
	ID          int64
	UserID      int64
	Title       string
	Description string
	Status      TaskStatus
	Priority    TaskPriority
	DueDate     *time.Time
	CategoryID  *int64
	Category    *Category
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Comments    []TaskComment
}

type Category struct {
	ID        int64
	UserID    int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TaskComment struct {
	ID        int64
	TaskID    int64
	UserID    int64
	Content   string
	CreatedAt time.Time
}
