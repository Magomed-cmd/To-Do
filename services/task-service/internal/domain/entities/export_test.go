package entities

import "testing"

func TestExportFormat_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		format ExportFormat
		want   bool
	}{
		{name: "csv is valid", format: ExportFormatCSV, want: true},
		{name: "ical is valid", format: ExportFormatICal, want: true},
		{name: "empty is invalid", format: "", want: false},
		{name: "unknown is invalid", format: "pdf", want: false},
		{name: "uppercase CSV is invalid", format: "CSV", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.format.IsValid(); got != tt.want {
				t.Errorf("ExportFormat.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExportFormat_String(t *testing.T) {
	tests := []struct {
		format ExportFormat
		want   string
	}{
		{format: ExportFormatCSV, want: "csv"},
		{format: ExportFormatICal, want: "ical"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.format.String(); got != tt.want {
				t.Errorf("ExportFormat.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExportFormat_ContentType(t *testing.T) {
	tests := []struct {
		format ExportFormat
		want   string
	}{
		{format: ExportFormatCSV, want: "text/csv; charset=utf-8"},
		{format: ExportFormatICal, want: "text/calendar; charset=utf-8"},
		{format: "unknown", want: "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			if got := tt.format.ContentType(); got != tt.want {
				t.Errorf("ExportFormat.ContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExportFormat_FileExtension(t *testing.T) {
	tests := []struct {
		format ExportFormat
		want   string
	}{
		{format: ExportFormatCSV, want: "csv"},
		{format: ExportFormatICal, want: "ics"},
		{format: "unknown", want: "bin"},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			if got := tt.format.FileExtension(); got != tt.want {
				t.Errorf("ExportFormat.FileExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}
