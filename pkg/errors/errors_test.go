package errors

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected string
	}{
		{
			name:     "message only",
			err:      New(CodeNotFound, "user not found"),
			expected: "user not found",
		},
		{
			name:     "message with details",
			err:      New(CodeNotFound, "user not found").WithDetails("id=123"),
			expected: "user not found: id=123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAppError_Is(t *testing.T) {
	err1 := ErrUserNotFound
	err2 := ErrUserNotFound.WithMessage("custom message")
	err3 := ErrTaskNotFound

	if !errors.Is(err1, err2) {
		t.Error("expected err1 and err2 to match (same code)")
	}

	if errors.Is(err1, err3) {
		t.Error("expected err1 and err3 NOT to match (different codes)")
	}
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrapped := Wrap(CodeInternal, originalErr, "wrapped message")

	if !errors.Is(wrapped, originalErr) {
		t.Error("expected Unwrap to return original error")
	}
}

func TestAppError_WithMethods(t *testing.T) {
	base := New(CodeValidation, "validation failed")

	withMsg := base.WithMessage("custom validation")
	if withMsg.Message != "custom validation" {
		t.Errorf("WithMessage failed: %s", withMsg.Message)
	}

	withDetails := base.WithDetails("field 'email' is invalid")
	if withDetails.Details != "field 'email' is invalid" {
		t.Errorf("WithDetails failed: %s", withDetails.Details)
	}

	cause := errors.New("root cause")
	withCause := base.WithCause(cause)
	if withCause.Cause != cause {
		t.Error("WithCause failed")
	}

	withMeta := base.WithMeta("field", "email")
	if withMeta.Metadata["field"] != "email" {
		t.Error("WithMeta failed")
	}
}

func TestCodeToHTTPStatus(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected int
	}{
		{CodeValidation, http.StatusBadRequest},
		{CodeBadRequest, http.StatusBadRequest},
		{CodeUnauthorized, http.StatusUnauthorized},
		{CodeInvalidCredentials, http.StatusUnauthorized},
		{CodeForbidden, http.StatusForbidden},
		{CodeNotFound, http.StatusNotFound},
		{CodeUserNotFound, http.StatusNotFound},
		{CodeTaskNotFound, http.StatusNotFound},
		{CodeAlreadyExists, http.StatusConflict},
		{CodeTooManyRequests, http.StatusTooManyRequests},
		{CodeServiceUnavailable, http.StatusServiceUnavailable},
		{CodeInternal, http.StatusInternalServerError},
		{"UNKNOWN_CODE", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			if got := CodeToHTTPStatus(tt.code); got != tt.expected {
				t.Errorf("CodeToHTTPStatus(%s) = %d, want %d", tt.code, got, tt.expected)
			}
		})
	}
}

func TestCodeToGRPCCode(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected codes.Code
	}{
		{CodeValidation, codes.InvalidArgument},
		{CodeUnauthorized, codes.Unauthenticated},
		{CodeForbidden, codes.PermissionDenied},
		{CodeNotFound, codes.NotFound},
		{CodeAlreadyExists, codes.AlreadyExists},
		{CodeInternal, codes.Internal},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			if got := CodeToGRPCCode(tt.code); got != tt.expected {
				t.Errorf("CodeToGRPCCode(%s) = %v, want %v", tt.code, got, tt.expected)
			}
		})
	}
}

func TestWriteHTTP(t *testing.T) {
	w := httptest.NewRecorder()
	err := ErrUserNotFound.WithDetails("id=42")

	WriteHTTP(w, err)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "USER_NOT_FOUND") {
		t.Errorf("expected body to contain error code, got: %s", body)
	}
	if !strings.Contains(body, "user not found") {
		t.Errorf("expected body to contain message, got: %s", body)
	}
}

func TestAsAppError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		if AsAppError(nil) != nil {
			t.Error("expected nil for nil error")
		}
	})

	t.Run("already AppError", func(t *testing.T) {
		original := ErrValidation.WithDetails("test")
		result := AsAppError(original)
		if result.Code != original.Code || result.Details != original.Details {
			t.Error("expected same AppError")
		}
	})

	t.Run("unknown error", func(t *testing.T) {
		result := AsAppError(errors.New("unknown"))
		if result.Code != CodeInternal {
			t.Errorf("expected CodeInternal, got %s", result.Code)
		}
	})
}

func TestIsHelpers(t *testing.T) {
	if !IsNotFound(ErrUserNotFound) {
		t.Error("IsNotFound should return true for ErrUserNotFound")
	}
	if !IsNotFound(ErrTaskNotFound) {
		t.Error("IsNotFound should return true for ErrTaskNotFound")
	}
	if IsNotFound(ErrValidation) {
		t.Error("IsNotFound should return false for ErrValidation")
	}

	if !IsUnauthorized(ErrInvalidCredentials) {
		t.Error("IsUnauthorized should return true for ErrInvalidCredentials")
	}
	if !IsUnauthorized(ErrTokenExpired) {
		t.Error("IsUnauthorized should return true for ErrTokenExpired")
	}

	if !IsForbidden(ErrInsufficientPrivileges) {
		t.Error("IsForbidden should return true for ErrInsufficientPrivileges")
	}

	if !IsValidation(ErrBadRequest) {
		t.Error("IsValidation should return true for ErrBadRequest")
	}
	if !IsValidation(ErrInvalidTaskStatus) {
		t.Error("IsValidation should return true for ErrInvalidTaskStatus")
	}
}

func TestIsCode(t *testing.T) {
	if !IsCode(ErrUserNotFound, CodeUserNotFound) {
		t.Error("IsCode should match")
	}
	if IsCode(ErrUserNotFound, CodeTaskNotFound) {
		t.Error("IsCode should not match different codes")
	}
	if IsCode(errors.New("plain error"), CodeInternal) {
		t.Error("IsCode should return false for non-AppError")
	}
}
