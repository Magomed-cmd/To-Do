package export

import (
	"todoapp/services/task-service/internal/domain/entities"
)

// Formatter defines the interface for task export formatters.
type Formatter interface {
	// Format converts a list of tasks to the target format.
	Format(tasks []entities.Task) ([]byte, error)
}

// NewFormatter creates a new formatter for the specified export format.
func NewFormatter(format entities.ExportFormat) (Formatter, error) {
	switch format {
	case entities.ExportFormatCSV:
		return NewCSVFormatter(), nil
	case entities.ExportFormatICal:
		return NewICalFormatter(), nil
	default:
		return nil, ErrUnsupportedFormat
	}
}
