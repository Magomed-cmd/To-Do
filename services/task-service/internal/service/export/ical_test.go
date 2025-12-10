package export

import (
	"strings"
	"testing"
	"time"

	"todoapp/services/task-service/internal/domain/entities"
)

func TestICalFormatter_Format_EmptyTasks(t *testing.T) {
	formatter := NewICalFormatter()

	data, err := formatter.Format([]entities.Task{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(data)

	// Check VCALENDAR structure
	if !strings.HasPrefix(content, "BEGIN:VCALENDAR\r\n") {
		t.Error("expected VCALENDAR header")
	}
	if !strings.HasSuffix(content, "END:VCALENDAR\r\n") {
		t.Error("expected VCALENDAR footer")
	}
	if !strings.Contains(content, "VERSION:2.0") {
		t.Error("expected VERSION:2.0")
	}
	if !strings.Contains(content, "PRODID:") {
		t.Error("expected PRODID")
	}

	// Should not contain any VTODO
	if strings.Contains(content, "BEGIN:VTODO") {
		t.Error("expected no VTODO for empty tasks")
	}
}

func TestICalFormatter_Format_SingleTask(t *testing.T) {
	formatter := NewICalFormatter()
	now := time.Date(2024, 12, 10, 10, 0, 0, 0, time.UTC)
	dueDate := time.Date(2024, 12, 15, 18, 0, 0, 0, time.UTC)

	tasks := []entities.Task{
		{
			ID:          42,
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

	content := string(data)

	// Check VTODO structure
	if !strings.Contains(content, "BEGIN:VTODO\r\n") {
		t.Error("expected VTODO begin")
	}
	if !strings.Contains(content, "END:VTODO\r\n") {
		t.Error("expected VTODO end")
	}

	// Check required fields
	if !strings.Contains(content, "UID:task-42@todoapp") {
		t.Error("expected UID")
	}
	if !strings.Contains(content, "SUMMARY:Test Task") {
		t.Error("expected SUMMARY")
	}
	if !strings.Contains(content, "DESCRIPTION:Test Description") {
		t.Error("expected DESCRIPTION")
	}
	if !strings.Contains(content, "PRIORITY:1") { // high = 1
		t.Error("expected PRIORITY:1 for high priority")
	}
	if !strings.Contains(content, "STATUS:NEEDS-ACTION") {
		t.Error("expected STATUS:NEEDS-ACTION for pending")
	}
	if !strings.Contains(content, "CATEGORIES:Work") {
		t.Error("expected CATEGORIES")
	}

	// Check dates
	if !strings.Contains(content, "DTSTAMP:20241210T100000Z") {
		t.Error("expected DTSTAMP")
	}
	if !strings.Contains(content, "DUE:20241215T180000Z") {
		t.Error("expected DUE date")
	}
}

func TestICalFormatter_Format_MultipleTasks(t *testing.T) {
	formatter := NewICalFormatter()
	now := time.Now()

	tasks := []entities.Task{
		{ID: 1, Title: "Task 1", Status: entities.TaskStatusPending, Priority: entities.TaskPriorityLow, CreatedAt: now, UpdatedAt: now},
		{ID: 2, Title: "Task 2", Status: entities.TaskStatusInProgress, Priority: entities.TaskPriorityMedium, CreatedAt: now, UpdatedAt: now},
		{ID: 3, Title: "Task 3", Status: entities.TaskStatusCompleted, Priority: entities.TaskPriorityHigh, CreatedAt: now, UpdatedAt: now},
	}

	data, err := formatter.Format(tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(data)

	// Count VTODO occurrences
	vtodoCount := strings.Count(content, "BEGIN:VTODO")
	if vtodoCount != 3 {
		t.Errorf("expected 3 VTODOs, got %d", vtodoCount)
	}

	// Check UIDs are unique
	if !strings.Contains(content, "UID:task-1@todoapp") {
		t.Error("expected UID for task 1")
	}
	if !strings.Contains(content, "UID:task-2@todoapp") {
		t.Error("expected UID for task 2")
	}
	if !strings.Contains(content, "UID:task-3@todoapp") {
		t.Error("expected UID for task 3")
	}
}

func TestICalFormatter_PriorityMapping(t *testing.T) {
	tests := []struct {
		priority entities.TaskPriority
		want     int
	}{
		{entities.TaskPriorityHigh, 1},
		{entities.TaskPriorityMedium, 5},
		{entities.TaskPriorityLow, 9},
		{"unknown", 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			if got := mapPriority(tt.priority); got != tt.want {
				t.Errorf("mapPriority(%s) = %d, want %d", tt.priority, got, tt.want)
			}
		})
	}
}

func TestICalFormatter_StatusMapping(t *testing.T) {
	tests := []struct {
		status entities.TaskStatus
		want   string
	}{
		{entities.TaskStatusPending, "NEEDS-ACTION"},
		{entities.TaskStatusInProgress, "IN-PROCESS"},
		{entities.TaskStatusCompleted, "COMPLETED"},
		{entities.TaskStatusArchived, "COMPLETED"},
		{"unknown", "NEEDS-ACTION"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := mapStatus(tt.status); got != tt.want {
				t.Errorf("mapStatus(%s) = %s, want %s", tt.status, got, tt.want)
			}
		})
	}
}

func TestICalFormatter_EscapeText(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"plain text", "plain text"},
		{"text with, comma", "text with\\, comma"},
		{"text with; semicolon", "text with\\; semicolon"},
		{"text with\nnewline", "text with\\nnewline"},
		{"text with\\backslash", "text with\\\\backslash"},
		{"mixed,;chars\n\\here", "mixed\\,\\;chars\\n\\\\here"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := escapeICalText(tt.input); got != tt.want {
				t.Errorf("escapeICalText(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestICalFormatter_FormatTime(t *testing.T) {
	// Test UTC conversion
	loc, _ := time.LoadLocation("America/New_York")
	localTime := time.Date(2024, 12, 10, 10, 30, 45, 0, loc)

	result := formatICalTime(localTime)

	// Should be in UTC format with Z suffix
	if !strings.HasSuffix(result, "Z") {
		t.Errorf("expected UTC time with Z suffix, got %s", result)
	}

	// Format should be YYYYMMDDTHHMMSSZ
	if len(result) != 16 {
		t.Errorf("expected 16 character format, got %d: %s", len(result), result)
	}
}

func TestICalFormatter_NilDueDateAndCategory(t *testing.T) {
	formatter := NewICalFormatter()
	now := time.Now()

	tasks := []entities.Task{
		{
			ID:        1,
			Title:     "Task without optional fields",
			Status:    entities.TaskStatusPending,
			Priority:  entities.TaskPriorityLow,
			DueDate:   nil,
			Category:  nil,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	data, err := formatter.Format(tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(data)

	// Should not contain DUE field
	if strings.Contains(content, "DUE:") {
		t.Error("expected no DUE field for nil due date")
	}

	// Should not contain CATEGORIES field
	if strings.Contains(content, "CATEGORIES:") {
		t.Error("expected no CATEGORIES field for nil category")
	}
}

func TestICalFormatter_EmptyDescription(t *testing.T) {
	formatter := NewICalFormatter()
	now := time.Now()

	tasks := []entities.Task{
		{
			ID:          1,
			Title:       "Task",
			Description: "",
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

	content := string(data)

	// Should not contain DESCRIPTION field for empty description
	if strings.Contains(content, "DESCRIPTION:") {
		t.Error("expected no DESCRIPTION field for empty description")
	}
}
