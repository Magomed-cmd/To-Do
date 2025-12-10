package export

import "errors"

var (
	// ErrUnsupportedFormat is returned when an unsupported export format is requested.
	ErrUnsupportedFormat = errors.New("unsupported export format")
)
