package export

import (
	"bytes"
	"encoding/csv"
	"strconv"
	"time"

	"todoapp/services/task-service/internal/domain/entities"
)

// CSVFormatter formats tasks as CSV.
type CSVFormatter struct{}

// NewCSVFormatter creates a new CSV formatter.
func NewCSVFormatter() *CSVFormatter {
	return &CSVFormatter{}
}

// Format converts tasks to CSV format.
// The output includes a UTF-8 BOM for proper Excel compatibility.
func (f *CSVFormatter) Format(tasks []entities.Task) ([]byte, error) {
	var buf bytes.Buffer

	// Write UTF-8 BOM for Excel compatibility
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"ID", "Title", "Description", "Status", "Priority", "DueDate", "Category", "CreatedAt", "UpdatedAt"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Write task rows
	for _, task := range tasks {
		row := f.taskToRow(task)
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (f *CSVFormatter) taskToRow(task entities.Task) []string {
	dueDate := ""
	if task.DueDate != nil {
		dueDate = task.DueDate.Format(time.RFC3339)
	}

	category := ""
	if task.Category != nil {
		category = task.Category.Name
	}

	return []string{
		strconv.FormatInt(task.ID, 10),
		task.Title,
		task.Description,
		string(task.Status),
		string(task.Priority),
		dueDate,
		category,
		task.CreatedAt.Format(time.RFC3339),
		task.UpdatedAt.Format(time.RFC3339),
	}
}
