package dto

import (
	"testing"
	"time"

	"todoapp/services/task-service/internal/domain/entities"
)

func TestCreateTaskRequest_ToInputDefaults(t *testing.T) {
	req := CreateTaskRequest{
		Title:       "  title ",
		Description: "  desc ",
	}

	input := req.ToInput(10)
	if input.UserID != 10 {
		t.Fatalf("unexpected user id: %d", input.UserID)
	}
	if input.Status != entities.TaskStatusPending {
		t.Fatalf("expected default status pending")
	}
	if input.Priority != entities.TaskPriorityMedium {
		t.Fatalf("expected default priority medium")
	}
	if input.Description != "desc" {
		t.Fatalf("expected trimmed description, got %q", input.Description)
	}
}

func TestUpdateTaskRequest_ToInput(t *testing.T) {
	title := "  New title "
	description := "  details "
	status := "COMPLETED"
	priority := "HIGH"
	dueStr := "2024-12-15T10:00:00Z"
	categoryID := int64(3)

	req := UpdateTaskRequest{
		Title:         &title,
		Description:   &description,
		Status:        &status,
		Priority:      &priority,
		DueDate:       &dueStr,
		ClearDueDate:  true,
		CategoryID:    &categoryID,
		ClearCategory: true,
	}

	input := req.ToInput(1, 2)
	if *input.Title != "  New title " {
		t.Fatalf("unexpected title pointer")
	}
	if *input.Description != "details" {
		t.Fatalf("expected trimmed description")
	}
	if input.Status == nil || *input.Status != entities.TaskStatusCompleted {
		t.Fatalf("unexpected status: %v", input.Status)
	}
	if input.Priority == nil || *input.Priority != entities.TaskPriorityHigh {
		t.Fatalf("unexpected priority: %v", input.Priority)
	}
	if input.DueDate == nil || !input.ClearDueDate || input.CategoryID == nil || *input.CategoryID != categoryID || !input.ClearCategory {
		t.Fatalf("unexpected struct: %+v", input)
	}
}

func TestTaskFilterRequest_ToFilter(t *testing.T) {
	fromStr := "2024-12-10T10:00:00Z"
	toStr := "2024-12-11T10:00:00Z"
	categoryID := int64(2)

	req := TaskFilterRequest{
		Status:     "pending",
		Priority:   "medium",
		CategoryID: &categoryID,
		Search:     "  hello ",
		DueFrom:    &fromStr,
		DueTo:      &toStr,
		Limit:      -1,
		Offset:     -2,
	}

	filter := req.ToFilter()
	if len(filter.Statuses) != 1 || filter.Statuses[0] != entities.TaskStatusPending {
		t.Fatalf("unexpected statuses: %+v", filter.Statuses)
	}
	if len(filter.Priorities) != 1 || filter.Priorities[0] != entities.TaskPriorityMedium {
		t.Fatalf("unexpected priorities: %+v", filter.Priorities)
	}
	if filter.CategoryID == nil || *filter.CategoryID != categoryID {
		t.Fatalf("unexpected category id")
	}
	if filter.Search != "hello" {
		t.Fatalf("expected trimmed search string")
	}
	if filter.Limit != 20 || filter.Offset != 0 {
		t.Fatalf("unexpected pagination: limit=%d offset=%d", filter.Limit, filter.Offset)
	}
	if filter.DueFrom == nil || filter.DueTo == nil {
		t.Fatalf("expected DueFrom and DueTo to be parsed")
	}
}

func TestResponses(t *testing.T) {
	now := time.Now()
	category := entities.Category{ID: 2, Name: "Work", CreatedAt: now, UpdatedAt: now}
	task := entities.Task{
		ID:          1,
		UserID:      5,
		Title:       "Test",
		Description: "Desc",
		Status:      entities.TaskStatusInProgress,
		Priority:    entities.TaskPriorityHigh,
		DueDate:     &now,
		CategoryID:  &category.ID,
		Category:    &category,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	taskResp := NewTaskResponse(task)
	if taskResp.Category == nil || taskResp.Category.ID != category.ID {
		t.Fatalf("expected embedded category")
	}

	cats := NewCategoryResponses([]entities.Category{category})
	if len(cats) != 1 || cats[0].ID != category.ID {
		t.Fatalf("unexpected categories response: %+v", cats)
	}

	comment := entities.TaskComment{ID: 3, TaskID: 1, UserID: 5, Content: "ok", CreatedAt: now}
	comments := NewCommentResponses([]entities.TaskComment{comment})
	if len(comments) != 1 || comments[0].Content != "ok" {
		t.Fatalf("unexpected comments response")
	}
}

func TestHelperFunctions(t *testing.T) {
	if status, ok := parseStatus(" archived "); !ok || status != entities.TaskStatusArchived {
		t.Fatalf("parseStatus failed")
	}
	if _, ok := parseStatus("invalid"); ok {
		t.Fatalf("parseStatus should fail for invalid")
	}

	if priority, ok := parsePriority("LOW"); !ok || priority != entities.TaskPriorityLow {
		t.Fatalf("parsePriority failed")
	}
	if _, ok := parsePriority("bad"); ok {
		t.Fatalf("parsePriority should fail for invalid")
	}

	if toStatusOrDefault("unknown") != entities.TaskStatusPending {
		t.Fatalf("expected default pending status")
	}
	if toPriorityOrDefault("unknown") != entities.TaskPriorityMedium {
		t.Fatalf("expected default medium priority")
	}

	str := " trimmed "
	ptr := normalizePtr(&str)
	if ptr == nil || *ptr != "trimmed" {
		t.Fatalf("normalizePtr failed")
	}
	if normalizePtr(nil) != nil {
		t.Fatalf("normalizePtr should return nil for nil input")
	}

	if clampLimit(0) != 20 || clampLimit(150) != 100 || clampLimit(10) != 10 {
		t.Fatalf("clampLimit unexpected")
	}
	if clampOffset(-1) != 0 || clampOffset(5) != 5 {
		t.Fatalf("clampOffset unexpected")
	}
}

func TestCreateCommentRequest_ToInput(t *testing.T) {
	req := CreateCommentRequest{Content: " hello "}
	input := req.ToInput(1, 2)
	if input.UserID != 1 || input.TaskID != 2 || input.Content != "hello" {
		t.Fatalf("unexpected input: %+v", input)
	}
}

func TestCreateCategoryRequest_ToInput(t *testing.T) {
	req := CreateCategoryRequest{Name: " Work "}
	input := req.ToInput(10)
	if input.UserID != 10 || input.Name != "Work" {
		t.Fatalf("unexpected category input: %+v", input)
	}
}
