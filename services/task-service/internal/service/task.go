package service

import (
	"context"
	"strings"
	"time"

	"todoapp/services/task-service/internal/domain"
	"todoapp/services/task-service/internal/domain/entities"
	"todoapp/services/task-service/internal/ports"
)

type TaskService struct {
	repo ports.TaskRepository
	now  func() time.Time
}

var _ ports.TaskService = (*TaskService)(nil)

func NewTaskService(repo ports.TaskRepository) *TaskService {
	return &TaskService{
		repo: repo,
		now:  time.Now,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, input ports.CreateTaskInput) (*entities.Task, error) {
	if err := s.validateTitle(input.Title); err != nil {
		return nil, err
	}

	if err := s.validateStatus(input.Status); err != nil {
		return nil, err
	}

	if err := s.validatePriority(input.Priority); err != nil {
		return nil, err
	}

	category, err := s.ensureCategory(ctx, input.UserID, input.CategoryID)
	if err != nil {
		return nil, err
	}

	task := &entities.Task{
		UserID:      input.UserID,
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		Status:      input.Status,
		Priority:    input.Priority,
		DueDate:     input.DueDate,
		CategoryID:  input.CategoryID,
		Category:    category,
	}

	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, input ports.UpdateTaskInput) (*entities.Task, error) {
	task, err := s.repo.GetTask(ctx, input.UserID, input.TaskID)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		if err := s.validateTitle(*input.Title); err != nil {
			return nil, err
		}
		task.Title = strings.TrimSpace(*input.Title)
	}

	if input.Description != nil {
		task.Description = strings.TrimSpace(*input.Description)
	}

	if input.Status != nil {
		task.Status = *input.Status
		if err := s.validateStatus(task.Status); err != nil {
			return nil, err
		}
	}

	if input.Priority != nil {
		task.Priority = *input.Priority
		if err := s.validatePriority(task.Priority); err != nil {
			return nil, err
		}
	}

	if input.DueDate != nil {
		task.DueDate = input.DueDate
	} else if input.ClearDueDate {
		task.DueDate = nil
	}

	if input.CategoryID != nil {
		category, err := s.ensureCategory(ctx, input.UserID, input.CategoryID)
		if err != nil {
			return nil, err
		}
		task.CategoryID = input.CategoryID
		task.Category = category
	} else if input.ClearCategory {
		task.CategoryID = nil
		task.Category = nil
	}

	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskService) UpdateTaskStatus(ctx context.Context, userID, taskID int64, status entities.TaskStatus) (*entities.Task, error) {
	if err := s.validateStatus(status); err != nil {
		return nil, err
	}

	task, err := s.repo.GetTask(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}

	task.Status = status

	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, userID, taskID int64) error {
	return s.repo.SoftDeleteTask(ctx, userID, taskID, s.now())
}

func (s *TaskService) GetTask(ctx context.Context, userID, taskID int64) (*entities.Task, error) {
	return s.repo.GetTask(ctx, userID, taskID)
}

func (s *TaskService) ListTasks(ctx context.Context, userID int64, filter ports.TaskFilter) ([]entities.Task, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.repo.ListTasks(ctx, userID, filter)
}

func (s *TaskService) CreateCategory(ctx context.Context, input ports.CreateCategoryInput) (*entities.Category, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, domain.ErrValidationFailed.WithMessage("category name is required")
	}

	category := &entities.Category{
		UserID: input.UserID,
		Name:   strings.TrimSpace(input.Name),
	}

	if err := s.repo.CreateCategory(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *TaskService) ListCategories(ctx context.Context, userID int64) ([]entities.Category, error) {
	return s.repo.ListCategories(ctx, userID)
}

func (s *TaskService) DeleteCategory(ctx context.Context, userID, categoryID int64) error {
	return s.repo.DeleteCategory(ctx, userID, categoryID)
}

func (s *TaskService) AddComment(ctx context.Context, input ports.AddCommentInput) (*entities.TaskComment, error) {
	if strings.TrimSpace(input.Content) == "" {
		return nil, domain.ErrValidationFailed.WithMessage("comment cannot be empty")
	}

	if _, err := s.repo.GetTask(ctx, input.UserID, input.TaskID); err != nil {
		return nil, err
	}

	comment := &entities.TaskComment{
		TaskID:  input.TaskID,
		UserID:  input.UserID,
		Content: strings.TrimSpace(input.Content),
	}

	if err := s.repo.CreateComment(ctx, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *TaskService) ListComments(ctx context.Context, userID, taskID int64) ([]entities.TaskComment, error) {
	if _, err := s.repo.GetTask(ctx, userID, taskID); err != nil {
		return nil, err
	}
	return s.repo.ListComments(ctx, userID, taskID)
}

func (s *TaskService) WithNow(now func() time.Time) {
	s.now = now
}

func (s *TaskService) ensureCategory(ctx context.Context, userID int64, categoryID *int64) (*entities.Category, error) {
	if categoryID == nil {
		return nil, nil
	}

	category, err := s.repo.GetCategory(ctx, userID, *categoryID)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (s *TaskService) validateTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return domain.ErrValidationFailed.WithMessage("title is required")
	}
	return nil
}

func (s *TaskService) validateStatus(status entities.TaskStatus) error {
	switch status {
	case entities.TaskStatusPending,
		entities.TaskStatusInProgress,
		entities.TaskStatusCompleted,
		entities.TaskStatusArchived:
		return nil
	default:
		return domain.ErrInvalidTaskStatus.WithMessage("unsupported status: " + string(status))
	}
}

func (s *TaskService) validatePriority(priority entities.TaskPriority) error {
	switch priority {
	case entities.TaskPriorityLow,
		entities.TaskPriorityMedium,
		entities.TaskPriorityHigh:
		return nil
	default:
		return domain.ErrInvalidTaskPriority.WithMessage("unsupported priority: " + string(priority))
	}
}
