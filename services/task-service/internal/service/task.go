package service

import (
	"context"
	"io"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"todoapp/pkg/events"
	analyticsv1 "todoapp/pkg/proto/analytics/v1"
	"todoapp/services/task-service/internal/domain"
	"todoapp/services/task-service/internal/domain/entities"
	"todoapp/services/task-service/internal/ports"
	"todoapp/services/task-service/internal/service/export"
)

type TaskService struct {
	repo      ports.TaskRepository
	users     ports.UserDirectory
	analytics ports.AnalyticsTracker
	publisher ports.TaskEventPublisher
	now       func() time.Time
	logger    *log.Logger
}

type TaskServiceOption func(*TaskService)

var _ ports.TaskService = (*TaskService)(nil)

func NewTaskService(repo ports.TaskRepository, opts ...TaskServiceOption) *TaskService {
	svc := &TaskService{
		repo:   repo,
		now:    time.Now,
		logger: log.New(io.Discard, "", 0),
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

func WithUserDirectory(users ports.UserDirectory) TaskServiceOption {
	return func(s *TaskService) {
		s.users = users
	}
}

func WithAnalyticsTracker(tracker ports.AnalyticsTracker) TaskServiceOption {
	return func(s *TaskService) {
		s.analytics = tracker
	}
}

func WithEventPublisher(publisher ports.TaskEventPublisher) TaskServiceOption {
	return func(s *TaskService) {
		s.publisher = publisher
	}
}

func WithLogger(logger *log.Logger) TaskServiceOption {
	return func(s *TaskService) {
		if logger != nil {
			s.logger = logger
		}
	}
}

func (s *TaskService) CreateTask(ctx context.Context, input ports.CreateTaskInput) (*entities.Task, error) {
	user, err := s.ensureUser(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

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

	s.trackAnalyticsEvent(ctx, ports.AnalyticsEvent{
		Type:       analyticsv1.TaskEventType_TASK_EVENT_TYPE_CREATED,
		UserID:     task.UserID,
		TaskID:     task.ID,
		Status:     string(task.Status),
		Priority:   string(task.Priority),
		OccurredAt: s.now(),
	})
	s.publishTaskNotification(ctx, events.TaskEventCreated, task, user)

	return task, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, input ports.UpdateTaskInput) (*entities.Task, error) {
	if _, err := s.ensureUser(ctx, input.UserID); err != nil {
		return nil, err
	}

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
	user, err := s.ensureUser(ctx, userID)
	if err != nil {
		return nil, err
	}
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

	if status == entities.TaskStatusCompleted {
		s.trackAnalyticsEvent(ctx, ports.AnalyticsEvent{
			Type:       analyticsv1.TaskEventType_TASK_EVENT_TYPE_COMPLETED,
			UserID:     task.UserID,
			TaskID:     task.ID,
			Status:     string(task.Status),
			Priority:   string(task.Priority),
			OccurredAt: s.now(),
		})
		s.publishTaskNotification(ctx, events.TaskEventCompleted, task, user)
	}

	return task, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, userID, taskID int64) error {
	user, err := s.ensureUser(ctx, userID)
	if err != nil {
		return err
	}

	task, err := s.repo.GetTask(ctx, userID, taskID)
	if err != nil {
		return err
	}

	if err := s.repo.SoftDeleteTask(ctx, userID, taskID, s.now()); err != nil {
		return err
	}

	s.trackAnalyticsEvent(ctx, ports.AnalyticsEvent{
		Type:       analyticsv1.TaskEventType_TASK_EVENT_TYPE_DELETED,
		UserID:     task.UserID,
		TaskID:     task.ID,
		Status:     string(task.Status),
		Priority:   string(task.Priority),
		OccurredAt: s.now(),
	})
	s.publishTaskNotification(ctx, events.TaskEventDeleted, task, user)

	return nil
}

func (s *TaskService) GetTask(ctx context.Context, userID, taskID int64) (*entities.Task, error) {
	if _, err := s.ensureUser(ctx, userID); err != nil {
		return nil, err
	}
	return s.repo.GetTask(ctx, userID, taskID)
}

func (s *TaskService) ListTasks(ctx context.Context, userID int64, filter ports.TaskFilter) ([]entities.Task, error) {
	if _, err := s.ensureUser(ctx, userID); err != nil {
		return nil, err
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.repo.ListTasks(ctx, userID, filter)
}

func (s *TaskService) CreateCategory(ctx context.Context, input ports.CreateCategoryInput) (*entities.Category, error) {
	if _, err := s.ensureUser(ctx, input.UserID); err != nil {
		return nil, err
	}
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
	if _, err := s.ensureUser(ctx, userID); err != nil {
		return nil, err
	}
	return s.repo.ListCategories(ctx, userID)
}

func (s *TaskService) DeleteCategory(ctx context.Context, userID, categoryID int64) error {
	if _, err := s.ensureUser(ctx, userID); err != nil {
		return err
	}
	return s.repo.DeleteCategory(ctx, userID, categoryID)
}

func (s *TaskService) AddComment(ctx context.Context, input ports.AddCommentInput) (*entities.TaskComment, error) {
	if _, err := s.ensureUser(ctx, input.UserID); err != nil {
		return nil, err
	}
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
	if _, err := s.ensureUser(ctx, userID); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetTask(ctx, userID, taskID); err != nil {
		return nil, err
	}
	return s.repo.ListComments(ctx, userID, taskID)
}

func (s *TaskService) ExportTasks(ctx context.Context, userID int64, format entities.ExportFormat) ([]byte, string, error) {
	if _, err := s.ensureUser(ctx, userID); err != nil {
		return nil, "", err
	}

	if !format.IsValid() {
		return nil, "", domain.ErrValidationFailed.WithMessage("unsupported export format: " + format.String())
	}

	// Fetch all tasks without pagination limit for export
	tasks, err := s.repo.ListTasks(ctx, userID, ports.TaskFilter{Limit: 10000})
	if err != nil {
		return nil, "", err
	}

	formatter, err := export.NewFormatter(format)
	if err != nil {
		return nil, "", domain.ErrValidationFailed.WithMessage(err.Error())
	}

	data, err := formatter.Format(tasks)
	if err != nil {
		return nil, "", err
	}

	filename := s.generateExportFilename(format)
	return data, filename, nil
}

func (s *TaskService) generateExportFilename(format entities.ExportFormat) string {
	timestamp := s.now().Format("2006-01-02")
	return "tasks_" + timestamp + "." + format.FileExtension()
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

func (s *TaskService) ensureUser(ctx context.Context, userID int64) (*ports.UserInfo, error) {
	if s.users == nil || userID == 0 {
		return nil, nil
	}
	user, err := s.users.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !user.Active {
		return nil, domain.ErrForbiddenTaskAccess
	}
	return user, nil
}

func (s *TaskService) trackAnalyticsEvent(ctx context.Context, event ports.AnalyticsEvent) {
	if s.analytics == nil {
		return
	}
	if err := s.analytics.TrackTaskEvent(ctx, event); err != nil {
		s.logError("analytics tracking failed: %v", err)
	}
}

func (s *TaskService) publishTaskNotification(ctx context.Context, eventType events.TaskEventType, task *entities.Task, user *ports.UserInfo) {
	if s.publisher == nil || task == nil || user == nil || user.Email == "" {
		return
	}

	payload := events.TaskEvent{
		ID:          uuid.NewString(),
		Type:        eventType,
		TaskID:      task.ID,
		UserID:      user.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		Priority:    string(task.Priority),
		DueDate:     task.DueDate,
		UserEmail:   user.Email,
		CreatedAt:   s.now(),
	}

	go func() {
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		if err := s.publisher.Publish(ctx, payload); err != nil {
			s.logError("notification publish failed: %v", err)
		}
	}()
}

func (s *TaskService) logError(format string, args ...any) {
	if s.logger == nil {
		return
	}
	s.logger.Printf(format, args...)
}
