package entities

// ExportFormat represents the format for task export.
type ExportFormat string

const (
	// ExportFormatCSV is the CSV export format.
	ExportFormatCSV ExportFormat = "csv"
	// ExportFormatICal is the iCalendar (RFC 5545) export format.
	ExportFormatICal ExportFormat = "ical"
)

// IsValid checks if the export format is valid.
func (f ExportFormat) IsValid() bool {
	switch f {
	case ExportFormatCSV, ExportFormatICal:
		return true
	default:
		return false
	}
}

// String returns the string representation of the export format.
func (f ExportFormat) String() string {
	return string(f)
}

// ContentType returns the MIME content type for the export format.
func (f ExportFormat) ContentType() string {
	switch f {
	case ExportFormatCSV:
		return "text/csv; charset=utf-8"
	case ExportFormatICal:
		return "text/calendar; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}

// FileExtension returns the file extension for the export format.
func (f ExportFormat) FileExtension() string {
	switch f {
	case ExportFormatCSV:
		return "csv"
	case ExportFormatICal:
		return "ics"
	default:
		return "bin"
	}
}
