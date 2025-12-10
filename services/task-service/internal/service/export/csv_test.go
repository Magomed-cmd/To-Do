package export

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"todoapp/services/task-service/internal/domain/entities"
)

func TestCSVFormatter_Format_EmptyTasks(t *testing.T) {
	formatter := NewCSVFormatter()

	data, err := formatter.Format([]entities.Task{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check BOM is present
	if !bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		t.Error("expected UTF-8 BOM at start of file")
	}

	// Check header is present
	content := string(data[3:]) // Skip BOM
	if !strings.HasPrefix(content, "ID,Title,Description,Status,Priority,DueDate,Category,CreatedAt,UpdatedAt") {
		t.Errorf("expected header row, got: %s", content)
	}

	// Should have only header (plus newline)
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line (header only), got %d lines", len(lines))
	}
}

func TestCSVFormatter_Format_SingleTask(t *testing.T) {
	formatter := NewCSVFormatter()
	now := time.Date(2024, 12, 10, 10, 0, 0, 0, time.UTC)
	dueDate := time.Date(2024, 12, 15, 18, 0, 0, 0, time.UTC)

	tasks := []entities.Task{
		{
			ID:          1,
			Title:       "Test Task",
			Description: "Test Description",
			Status:      entities.TaskStatusPending,
			Priority:    entities.TaskPriorityHigh,
			DueDate:     &dueDate,
			Category:    &entities.Category{Name: "Work"},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	data, err := formatter.Format(tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(data[3:]) // Skip BOM
	lines := strings.Split(strings.TrimSpace(content), "\n")

	if len(lines) != 2 {
		t.Errorf("expected 2 lines (header + 1 task), got %d", len(lines))
	}

	// Parse CSV to verify content
	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	row := records[1]
	if row[0] != "1" {
		t.Errorf("expected ID=1, got %s", row[0])
	}
	if row[1] != "Test Task" {
		t.Errorf("expected Title='Test Task', got %s", row[1])
	}
	if row[3] != "pending" {
		t.Errorf("expected Status='pending', got %s", row[3])
	}
	if row[4] != "high" {
		t.Errorf("expected Priority='high', got %s", row[4])
	}
	if row[6] != "Work" {
		t.Errorf("expected Category='Work', got %s", row[6])
	}
}

func TestCSVFormatter_Format_MultipleTasks(t *testing.T) {
	formatter := NewCSVFormatter()
	now := time.Date(2024, 12, 10, 10, 0, 0, 0, time.UTC)

	tasks := []entities.Task{
		{
			ID:        1,
			Title:     "Task 1",
			Status:    entities.TaskStatusPending,
			Priority:  entities.TaskPriorityLow,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        2,
			Title:     "Task 2",
			Status:    entities.TaskStatusCompleted,
			Priority:  entities.TaskPriorityMedium,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        3,
			Title:     "Task 3",
			Status:    entities.TaskStatusInProgress,
			Priority:  entities.TaskPriorityHigh,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	data, err := formatter.Format(tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(data[3:]) // Skip BOM
	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	if len(records) != 4 {
		t.Errorf("expected 4 records (header + 3 tasks), got %d", len(records))
	}
}

func TestCSVFormatter_Format_SpecialCharacters(t *testing.T) {
	formatter := NewCSVFormatter()
	now := time.Now()

	tasks := []entities.Task{
		{
			ID:          1,
			Title:       "Task with, comma",
			Description: "Description with \"quotes\" and\nnewlines",
			Status:      entities.TaskStatusPending,
			Priority:    entities.TaskPriorityLow,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	data, err := formatter.Format(tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(data[3:]) // Skip BOM
	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV with special characters: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	row := records[1]
	if row[1] != "Task with, comma" {
		t.Errorf("expected title with comma to be preserved, got: %s", row[1])
	}
	if !strings.Contains(row[2], "quotes") {
		t.Errorf("expected description with quotes to be preserved, got: %s", row[2])
	}
}

func TestCSVFormatter_Format_NilDueDateAndCategory(t *testing.T) {
	formatter := NewCSVFormatter()
	now := time.Now()

	tasks := []entities.Task{
		{
			ID:          1,
			Title:       "Task without due date",
			Status:      entities.TaskStatusPending,
			Priority:    entities.TaskPriorityLow,
			DueDate:     nil,
			Category:    nil,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	data, err := formatter.Format(tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(data[3:]) // Skip BOM
	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	row := records[1]
	if row[5] != "" {
		t.Errorf("expected empty due date, got: %s", row[5])
	}
	if row[6] != "" {
		t.Errorf("expected empty category, got: %s", row[6])
	}
}
