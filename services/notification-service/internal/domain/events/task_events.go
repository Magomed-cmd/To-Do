package events

import "time"

type TaskEventType string

const (
	TaskCreated   TaskEventType = "task.created"
	TaskCompleted TaskEventType = "task.completed"
	TaskDeleted   TaskEventType = "task.deleted"
	TaskOverdue   TaskEventType = "task.overdue"
)

type TaskEvent struct {
	ID          string        `json:"id"`
	Type        TaskEventType `json:"type"`
	TaskID      int64         `json:"taskId"`
	UserID      int64         `json:"userId"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Status      string        `json:"status"`
	Priority    string        `json:"priority"`
	DueDate     *time.Time    `json:"dueDate"`
	UserEmail   string        `json:"userEmail"`
	CreatedAt   time.Time     `json:"createdAt"`
}

func (e TaskEvent) IsZero() bool {
	return e.ID == "" || e.Type == ""
}
