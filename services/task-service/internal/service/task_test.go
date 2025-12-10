package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"todoapp/pkg/events"
	analyticsv1 "todoapp/pkg/proto/analytics/v1"
	"todoapp/services/task-service/internal/domain"
	"todoapp/services/task-service/internal/domain/entities"
	"todoapp/services/task-service/internal/ports"
)

type repoMock struct {
	createdTask   *entities.Task
	createErr     error
	storedTask    *entities.Task
	getErr        error
	updateErr     error
	softDeleteErr error
	listFilter    ports.TaskFilter
	listResult    []entities.Task
	listErr       error

	category     *entities.Category
	categoryErr  error
	categories   []entities.Category
	categoryList error
	deleteCatErr error

	comment      *entities.TaskComment
	commentErr   error
	comments     []entities.TaskComment
	commentsErr  error
}

func (r *repoMock) CreateTask(ctx context.Context, task *entities.Task) error {
	r.createdTask = task
	task.ID = 1
	task.CreatedAt = time.Now()
	task.UpdatedAt = task.CreatedAt
	return r.createErr
}

func (r *repoMock) UpdateTask(ctx context.Context, task *entities.Task) error {
	r.storedTask = task
	task.UpdatedAt = time.Now()
	return r.updateErr
}

func (r *repoMock) SoftDeleteTask(ctx context.Context, userID, taskID int64, deletedAt time.Time) error {
	r.storedTask = &entities.Task{ID: taskID, UserID: userID}
	return r.softDeleteErr
}

func (r *repoMock) GetTask(ctx context.Context, userID, taskID int64) (*entities.Task, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if r.storedTask != nil {
		return r.storedTask, nil
	}
	return &entities.Task{ID: taskID, UserID: userID, Status: entities.TaskStatusPending, Priority: entities.TaskPriorityMedium}, nil
}

func (r *repoMock) ListTasks(ctx context.Context, userID int64, filter ports.TaskFilter) ([]entities.Task, error) {
	r.listFilter = filter
	return r.listResult, r.listErr
}

func (r *repoMock) CreateCategory(ctx context.Context, category *entities.Category) error {
	r.category = category
	category.ID = 2
	category.CreatedAt = time.Now()
	category.UpdatedAt = category.CreatedAt
	return nil
}

func (r *repoMock) ListCategories(ctx context.Context, userID int64) ([]entities.Category, error) {
	return r.categories, r.categoryList
}

func (r *repoMock) GetCategory(ctx context.Context, userID, categoryID int64) (*entities.Category, error) {
	return r.category, r.categoryErr
}

func (r *repoMock) DeleteCategory(ctx context.Context, userID, categoryID int64) error {
	return r.deleteCatErr
}

func (r *repoMock) CreateComment(ctx context.Context, comment *entities.TaskComment) error {
	r.comment = comment
	comment.ID = 3
	comment.CreatedAt = time.Now()
	return r.commentErr
}

func (r *repoMock) ListComments(ctx context.Context, userID, taskID int64) ([]entities.TaskComment, error) {
	return r.comments, r.commentsErr
}

type userDirStub struct {
	user *ports.UserInfo
	err  error
}

func (u userDirStub) GetUser(ctx context.Context, userID int64) (*ports.UserInfo, error) {
	return u.user, u.err
}

type analyticsStub struct {
	ch  chan ports.AnalyticsEvent
	err error
}

func (a analyticsStub) TrackTaskEvent(ctx context.Context, event ports.AnalyticsEvent) error {
	if a.ch != nil {
		a.ch <- event
	}
	return a.err
}

type publisherStub struct {
	ch  chan events.TaskEvent
	err error
}

func (p publisherStub) Publish(ctx context.Context, event events.TaskEvent) error {
	if p.ch != nil {
		p.ch <- event
	}
	return p.err
}

func TestCreateTaskSuccess(t *testing.T) {
	repo := &repoMock{
		category: &entities.Category{ID: 10, UserID: 1, Name: "Work"},
	}
	analyticsCh := make(chan ports.AnalyticsEvent, 1)
	publishCh := make(chan events.TaskEvent, 1)
	svc := NewTaskService(
		repo,
		WithUserDirectory(userDirStub{user: &ports.UserInfo{ID: 1, Email: "u@example.com", Active: true}}),
		WithAnalyticsTracker(analyticsStub{ch: analyticsCh}),
		WithEventPublisher(publisherStub{ch: publishCh}),
	)
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	svc.WithNow(func() time.Time { return now })

	categoryID := int64(10)
	task, err := svc.CreateTask(context.Background(), ports.CreateTaskInput{
		UserID:     1,
		Title:      "  title ",
		Priority:   entities.TaskPriorityHigh,
		Status:     entities.TaskStatusPending,
		CategoryID: &categoryID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "title" || task.Priority != entities.TaskPriorityHigh {
		t.Fatalf("task was not normalized: %+v", task)
	}
	if repo.createdTask == nil || repo.createdTask.Category == nil || repo.createdTask.Category.ID != 10 {
		t.Fatalf("task not persisted with category")
	}

	select {
	case ev := <-analyticsCh:
		if ev.Type != analyticsv1.TaskEventType_TASK_EVENT_TYPE_CREATED || ev.UserID != 1 {
			t.Fatalf("unexpected analytics event: %+v", ev)
		}
	case <-time.After(time.Second):
		t.Fatalf("analytics event not published")
	}

	select {
	case ev := <-publishCh:
		if ev.Type != events.TaskEventCreated || ev.UserEmail == "" {
			t.Fatalf("unexpected notification payload: %+v", ev)
		}
	case <-time.After(time.Second):
		t.Fatalf("notification not published")
	}
}

func TestCreateTaskValidation(t *testing.T) {
	repo := &repoMock{}
	svc := NewTaskService(repo)

	tests := []struct {
		name  string
		input ports.CreateTaskInput
	}{
		{name: "empty title", input: ports.CreateTaskInput{}},
		{name: "bad status", input: ports.CreateTaskInput{Title: "t", Status: "bad"}},
		{name: "bad priority", input: ports.CreateTaskInput{Title: "t", Status: entities.TaskStatusPending, Priority: "bad"}},
	}

	for _, tt := range tests {
		if _, err := svc.CreateTask(context.Background(), tt.input); err == nil {
			t.Fatalf("%s: expected error", tt.name)
		}
	}
}

func TestUpdateTask(t *testing.T) {
	repo := &repoMock{
		storedTask: &entities.Task{ID: 1, UserID: 1, Title: "old", Priority: entities.TaskPriorityMedium},
		category:   &entities.Category{ID: 2, UserID: 1, Name: "New"},
	}
	title := " new "
	priority := entities.TaskPriorityLow
	status := entities.TaskStatusInProgress
	clearDue := true
	categoryID := int64(2)

	svc := NewTaskService(repo, WithUserDirectory(userDirStub{user: &ports.UserInfo{ID: 1, Active: true}}))
	task, err := svc.UpdateTask(context.Background(), ports.UpdateTaskInput{
		UserID:       1,
		TaskID:       1,
		Title:        &title,
		Status:       &status,
		Priority:     &priority,
		ClearDueDate: clearDue,
		CategoryID:   &categoryID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "new" || task.Priority != priority || task.Status != status || task.CategoryID == nil || *task.CategoryID != 2 {
		t.Fatalf("task not updated: %+v", task)
	}
}

func TestUpdateTaskStatus(t *testing.T) {
	repo := &repoMock{
		storedTask: &entities.Task{ID: 1, UserID: 1, Status: entities.TaskStatusPending, Priority: entities.TaskPriorityHigh},
	}
	analyticsCh := make(chan ports.AnalyticsEvent, 1)
	publishCh := make(chan events.TaskEvent, 1)
	svc := NewTaskService(
		repo,
		WithUserDirectory(userDirStub{user: &ports.UserInfo{ID: 1, Email: "a@b.c", Active: true}}),
		WithAnalyticsTracker(analyticsStub{ch: analyticsCh}),
		WithEventPublisher(publisherStub{ch: publishCh}),
	)
	svc.WithNow(func() time.Time { return time.Time{} })

	task, err := svc.UpdateTaskStatus(context.Background(), 1, 1, entities.TaskStatusCompleted)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Status != entities.TaskStatusCompleted {
		t.Fatalf("status not updated")
	}

	select {
	case <-analyticsCh:
	case <-time.After(time.Second):
		t.Fatalf("expected analytics event")
	}
	select {
	case <-publishCh:
	case <-time.After(time.Second):
		t.Fatalf("expected publish event")
	}
}

func TestDeleteTask(t *testing.T) {
	repo := &repoMock{storedTask: &entities.Task{ID: 1, UserID: 1, Status: entities.TaskStatusPending}}
	analyticsCh := make(chan ports.AnalyticsEvent, 1)
	publishCh := make(chan events.TaskEvent, 1)
	svc := NewTaskService(
		repo,
		WithUserDirectory(userDirStub{user: &ports.UserInfo{ID: 1, Email: "u@example.com", Active: true}}),
		WithAnalyticsTracker(analyticsStub{ch: analyticsCh}),
		WithEventPublisher(publisherStub{ch: publishCh}),
	)

	if err := svc.DeleteTask(context.Background(), 1, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case <-analyticsCh:
	case <-time.After(time.Second):
		t.Fatalf("analytics event not sent")
	}
	select {
	case <-publishCh:
	case <-time.After(time.Second):
		t.Fatalf("publish event not sent")
	}
}

func TestListTasksDefaults(t *testing.T) {
	repo := &repoMock{listResult: []entities.Task{{ID: 1, UserID: 1}}}
	svc := NewTaskService(repo, WithUserDirectory(userDirStub{user: &ports.UserInfo{ID: 1, Active: true}}))

	_, err := svc.ListTasks(context.Background(), 1, ports.TaskFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.listFilter.Limit != 20 || repo.listFilter.Offset != 0 {
		t.Fatalf("defaults not applied: %+v", repo.listFilter)
	}
}

func TestCategoryAndComments(t *testing.T) {
	repo := &repoMock{}
	svc := NewTaskService(repo, WithUserDirectory(userDirStub{user: &ports.UserInfo{ID: 1, Active: true}}))

	if _, err := svc.CreateCategory(context.Background(), ports.CreateCategoryInput{UserID: 1, Name: ""}); err == nil {
		t.Fatalf("expected validation error for empty category name")
	}

	_, err := svc.CreateCategory(context.Background(), ports.CreateCategoryInput{UserID: 1, Name: "Work"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	repo.storedTask = &entities.Task{ID: 1, UserID: 1}
	if _, err := svc.AddComment(context.Background(), ports.AddCommentInput{UserID: 1, TaskID: 1, Content: "  "}); err == nil {
		t.Fatalf("expected validation error for empty comment")
	}
	if _, err := svc.AddComment(context.Background(), ports.AddCommentInput{UserID: 1, TaskID: 1, Content: "ok"}); err != nil {
		t.Fatalf("unexpected error adding comment: %v", err)
	}
}

func TestEnsureUserInactive(t *testing.T) {
	svc := NewTaskService(&repoMock{}, WithUserDirectory(userDirStub{user: &ports.UserInfo{ID: 1, Active: false}}))
	if _, err := svc.GetTask(context.Background(), 1, 1); !errors.Is(err, domain.ErrForbiddenTaskAccess) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}
