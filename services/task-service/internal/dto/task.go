package dto

import (
	"strings"
	"time"

	"todoapp/services/task-service/internal/domain/entities"
	"todoapp/services/task-service/internal/ports"
)

type CreateTaskRequest struct {
	Title       string     `json:"title" binding:"required,min=1,max=200"`
	Description string     `json:"description" binding:"omitempty,max=2000"`
	Status      string     `json:"status" binding:"omitempty,oneof=pending in_progress completed archived"`
	Priority    string     `json:"priority" binding:"omitempty,oneof=low medium high"`
	DueDate     *time.Time `json:"dueDate" binding:"omitempty"`
	CategoryID  *int64     `json:"categoryId" binding:"omitempty,gte=1"`
}

type UpdateTaskRequest struct {
	Title         *string    `json:"title" binding:"omitempty,min=1,max=200"`
	Description   *string    `json:"description" binding:"omitempty,max=2000"`
	Status        *string    `json:"status" binding:"omitempty,oneof=pending in_progress completed archived"`
	Priority      *string    `json:"priority" binding:"omitempty,oneof=low medium high"`
	DueDate       *time.Time `json:"dueDate"`
	ClearDueDate  bool       `json:"clearDueDate"`
	CategoryID    *int64     `json:"categoryId" binding:"omitempty,gte=1"`
	ClearCategory bool       `json:"clearCategory"`
}

type UpdateTaskStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending in_progress completed archived"`
}

type TaskFilterRequest struct {
	Status     string     `form:"status"`
	Priority   string     `form:"priority"`
	CategoryID *int64     `form:"categoryId"`
	Search     string     `form:"search"`
	DueFrom    *time.Time `form:"dueFrom"`
	DueTo      *time.Time `form:"dueTo"`
	Limit      int        `form:"limit,default=20"`
	Offset     int        `form:"offset,default=0"`
}

type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required,min=1,max=120"`
}

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000"`
}

type TaskResponse struct {
	ID          int64          `json:"id"`
	UserID      int64          `json:"userId"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Status      string         `json:"status"`
	Priority    string         `json:"priority"`
	DueDate     *time.Time     `json:"dueDate,omitempty"`
	CategoryID  *int64         `json:"categoryId,omitempty"`
	Category    *CategoryShort `json:"category,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

type CategoryResponse struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"userId"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CategoryShort struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CommentResponse struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"taskId"`
	UserID    int64     `json:"userId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

func (r CreateTaskRequest) ToInput(userID int64) ports.CreateTaskInput {
	return ports.CreateTaskInput{
		UserID:      userID,
		Title:       r.Title,
		Description: strings.TrimSpace(r.Description),
		Status:      toStatusOrDefault(r.Status),
		Priority:    toPriorityOrDefault(r.Priority),
		DueDate:     r.DueDate,
		CategoryID:  r.CategoryID,
	}
}

func (r UpdateTaskRequest) ToInput(userID, taskID int64) ports.UpdateTaskInput {
	var status *entities.TaskStatus
	var priority *entities.TaskPriority

	if r.Status != nil {
		value := entities.TaskStatus(strings.ToLower(strings.TrimSpace(*r.Status)))
		status = &value
	}

	if r.Priority != nil {
		value := entities.TaskPriority(strings.ToLower(strings.TrimSpace(*r.Priority)))
		priority = &value
	}

	return ports.UpdateTaskInput{
		UserID:        userID,
		TaskID:        taskID,
		Title:         r.Title,
		Description:   normalizePtr(r.Description),
		Status:        status,
		Priority:      priority,
		DueDate:       r.DueDate,
		ClearDueDate:  r.ClearDueDate,
		CategoryID:    r.CategoryID,
		ClearCategory: r.ClearCategory,
	}
}

func (r UpdateTaskStatusRequest) ToStatus() entities.TaskStatus {
	return toStatusOrDefault(r.Status)
}

func (r CreateCategoryRequest) ToInput(userID int64) ports.CreateCategoryInput {
	return ports.CreateCategoryInput{
		UserID: userID,
		Name:   strings.TrimSpace(r.Name),
	}
}

func (r CreateCommentRequest) ToInput(userID, taskID int64) ports.AddCommentInput {
	return ports.AddCommentInput{
		UserID:  userID,
		TaskID:  taskID,
		Content: strings.TrimSpace(r.Content),
	}
}

func (r TaskFilterRequest) ToFilter() ports.TaskFilter {
	var (
		statuses   []entities.TaskStatus
		priorities []entities.TaskPriority
	)

	if statusValue, ok := parseStatus(r.Status); ok {
		statuses = append(statuses, statusValue)
	}

	if priorityValue, ok := parsePriority(r.Priority); ok {
		priorities = append(priorities, priorityValue)
	}

	return ports.TaskFilter{
		Statuses:   statuses,
		Priorities: priorities,
		CategoryID: r.CategoryID,
		Search:     strings.TrimSpace(r.Search),
		DueFrom:    r.DueFrom,
		DueTo:      r.DueTo,
		Limit:      clampLimit(r.Limit),
		Offset:     clampOffset(r.Offset),
	}
}

func NewTaskResponse(task entities.Task) TaskResponse {
	var category *CategoryShort

	if task.Category != nil {
		category = &CategoryShort{
			ID:   task.Category.ID,
			Name: task.Category.Name,
		}
	}

	return TaskResponse{
		ID:          task.ID,
		UserID:      task.UserID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		Priority:    string(task.Priority),
		DueDate:     task.DueDate,
		CategoryID:  task.CategoryID,
		Category:    category,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

func NewTaskResponses(tasks []entities.Task) []TaskResponse {
	result := make([]TaskResponse, 0, len(tasks))

	for _, task := range tasks {
		result = append(result, NewTaskResponse(task))
	}

	return result
}

func NewCategoryResponse(category entities.Category) CategoryResponse {
	return CategoryResponse{
		ID:        category.ID,
		UserID:    category.UserID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}

func NewCategoryResponses(categories []entities.Category) []CategoryResponse {
	result := make([]CategoryResponse, 0, len(categories))

	for _, category := range categories {
		result = append(result, NewCategoryResponse(category))
	}

	return result
}

func NewCommentResponse(comment entities.TaskComment) CommentResponse {
	return CommentResponse{
		ID:        comment.ID,
		TaskID:    comment.TaskID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}
}

func NewCommentResponses(comments []entities.TaskComment) []CommentResponse {
	result := make([]CommentResponse, 0, len(comments))

	for _, comment := range comments {
		result = append(result, NewCommentResponse(comment))
	}

	return result
}

func toStatusOrDefault(raw string) entities.TaskStatus {
	if status, ok := parseStatus(raw); ok {
		return status
	}
	return entities.TaskStatusPending
}

func toPriorityOrDefault(raw string) entities.TaskPriority {
	if priority, ok := parsePriority(raw); ok {
		return priority
	}
	return entities.TaskPriorityMedium
}

func parseStatus(raw string) (entities.TaskStatus, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(entities.TaskStatusPending):
		return entities.TaskStatusPending, true
	case string(entities.TaskStatusInProgress):
		return entities.TaskStatusInProgress, true
	case string(entities.TaskStatusCompleted):
		return entities.TaskStatusCompleted, true
	case string(entities.TaskStatusArchived):
		return entities.TaskStatusArchived, true
	default:
		return "", false
	}
}

func parsePriority(raw string) (entities.TaskPriority, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(entities.TaskPriorityLow):
		return entities.TaskPriorityLow, true
	case string(entities.TaskPriorityMedium):
		return entities.TaskPriorityMedium, true
	case string(entities.TaskPriorityHigh):
		return entities.TaskPriorityHigh, true
	default:
		return "", false
	}
}

func normalizePtr(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	copy := trimmed
	return &copy
}

func clampLimit(limit int) int {
	switch {
	case limit <= 0:
		return 20
	case limit > 100:
		return 100
	default:
		return limit
	}
}

func clampOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}
