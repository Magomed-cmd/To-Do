package export

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"todoapp/services/task-service/internal/domain/entities"
)

// ICalFormatter formats tasks as iCalendar (RFC 5545) VTODO components.
type ICalFormatter struct{}

// NewICalFormatter creates a new iCal formatter.
func NewICalFormatter() *ICalFormatter {
	return &ICalFormatter{}
}

// Format converts tasks to iCalendar format with VTODO components.
func (f *ICalFormatter) Format(tasks []entities.Task) ([]byte, error) {
	var buf bytes.Buffer

	// Write VCALENDAR header
	buf.WriteString("BEGIN:VCALENDAR\r\n")
	buf.WriteString("VERSION:2.0\r\n")
	buf.WriteString("PRODID:-//TodoApp//Task Export//EN\r\n")
	buf.WriteString("CALSCALE:GREGORIAN\r\n")
	buf.WriteString("METHOD:PUBLISH\r\n")

	// Write each task as VTODO
	for _, task := range tasks {
		f.writeVTodo(&buf, task)
	}

	// Write VCALENDAR footer
	buf.WriteString("END:VCALENDAR\r\n")

	return buf.Bytes(), nil
}

func (f *ICalFormatter) writeVTodo(buf *bytes.Buffer, task entities.Task) {
	buf.WriteString("BEGIN:VTODO\r\n")

	// UID - unique identifier
	buf.WriteString(fmt.Sprintf("UID:task-%d@todoapp\r\n", task.ID))

	// DTSTAMP - creation timestamp (required)
	buf.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", formatICalTime(task.CreatedAt)))

	// DTSTART - task creation date
	buf.WriteString(fmt.Sprintf("DTSTART:%s\r\n", formatICalTime(task.CreatedAt)))

	// DUE - due date if set
	if task.DueDate != nil {
		buf.WriteString(fmt.Sprintf("DUE:%s\r\n", formatICalTime(*task.DueDate)))
	}

	// SUMMARY - task title
	buf.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICalText(task.Title)))

	// DESCRIPTION - task description
	if task.Description != "" {
		buf.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICalText(task.Description)))
	}

	// PRIORITY - 1=high, 5=medium, 9=low
	buf.WriteString(fmt.Sprintf("PRIORITY:%d\r\n", mapPriority(task.Priority)))

	// STATUS - NEEDS-ACTION, IN-PROCESS, COMPLETED
	buf.WriteString(fmt.Sprintf("STATUS:%s\r\n", mapStatus(task.Status)))

	// CATEGORIES - category name if set
	if task.Category != nil {
		buf.WriteString(fmt.Sprintf("CATEGORIES:%s\r\n", escapeICalText(task.Category.Name)))
	}

	// LAST-MODIFIED
	buf.WriteString(fmt.Sprintf("LAST-MODIFIED:%s\r\n", formatICalTime(task.UpdatedAt)))

	// SEQUENCE - version number
	buf.WriteString("SEQUENCE:0\r\n")

	buf.WriteString("END:VTODO\r\n")
}

// formatICalTime formats a time.Time to iCalendar format (UTC).
func formatICalTime(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}

// escapeICalText escapes special characters in iCalendar text values.
func escapeICalText(s string) string {
	// Escape backslashes first
	s = strings.ReplaceAll(s, "\\", "\\\\")
	// Escape semicolons
	s = strings.ReplaceAll(s, ";", "\\;")
	// Escape commas
	s = strings.ReplaceAll(s, ",", "\\,")
	// Escape newlines
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

// mapPriority maps task priority to iCalendar priority (1-9).
// RFC 5545: 1-4 = high, 5 = medium, 6-9 = low
func mapPriority(p entities.TaskPriority) int {
	switch p {
	case entities.TaskPriorityHigh:
		return 1
	case entities.TaskPriorityMedium:
		return 5
	case entities.TaskPriorityLow:
		return 9
	default:
		return 0 // undefined
	}
}

// mapStatus maps task status to iCalendar VTODO status.
func mapStatus(s entities.TaskStatus) string {
	switch s {
	case entities.TaskStatusPending:
		return "NEEDS-ACTION"
	case entities.TaskStatusInProgress:
		return "IN-PROCESS"
	case entities.TaskStatusCompleted:
		return "COMPLETED"
	case entities.TaskStatusArchived:
		return "COMPLETED"
	default:
		return "NEEDS-ACTION"
	}
}

// PERCENT-COMPLETE helper for future use
func mapPercentComplete(s entities.TaskStatus) string {
	switch s {
	case entities.TaskStatusCompleted, entities.TaskStatusArchived:
		return strconv.Itoa(100)
	case entities.TaskStatusInProgress:
		return strconv.Itoa(50)
	default:
		return strconv.Itoa(0)
	}
}
