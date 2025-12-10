package export

import (
	"errors"
	"testing"

	"todoapp/services/task-service/internal/domain/entities"
)

func TestNewFormatter_CSV(t *testing.T) {
	formatter, err := NewFormatter(entities.ExportFormatCSV)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if formatter == nil {
		t.Fatal("expected non-nil formatter")
	}

	// Verify it's a CSV formatter
	if _, ok := formatter.(*CSVFormatter); !ok {
		t.Error("expected *CSVFormatter")
	}
}

func TestNewFormatter_ICal(t *testing.T) {
	formatter, err := NewFormatter(entities.ExportFormatICal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if formatter == nil {
		t.Fatal("expected non-nil formatter")
	}

	// Verify it's an iCal formatter
	if _, ok := formatter.(*ICalFormatter); !ok {
		t.Error("expected *ICalFormatter")
	}
}

func TestNewFormatter_UnsupportedFormat(t *testing.T) {
	formats := []entities.ExportFormat{
		"",
		"pdf",
		"json",
		"xml",
		"unknown",
	}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			formatter, err := NewFormatter(format)

			if formatter != nil {
				t.Error("expected nil formatter for unsupported format")
			}

			if !errors.Is(err, ErrUnsupportedFormat) {
				t.Errorf("expected ErrUnsupportedFormat, got %v", err)
			}
		})
	}
}

func TestFormatter_Interface(t *testing.T) {
	// Verify both formatters implement Formatter interface
	var _ Formatter = (*CSVFormatter)(nil)
	var _ Formatter = (*ICalFormatter)(nil)
}
