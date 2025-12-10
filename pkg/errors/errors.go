// Package errors provides a unified error handling system for all microservices.
// It includes structured error types, error codes, and automatic HTTP/gRPC status mapping.
package errors

import (
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
)

// ErrorCode represents a machine-readable error code.
type ErrorCode string

// Standard error codes used across all services.
const (
	// Generic errors
	CodeInternal           ErrorCode = "INTERNAL_ERROR"
	CodeValidation         ErrorCode = "VALIDATION_FAILED"
	CodeNotFound           ErrorCode = "NOT_FOUND"
	CodeAlreadyExists      ErrorCode = "ALREADY_EXISTS"
	CodeForbidden          ErrorCode = "FORBIDDEN"
	CodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	CodeBadRequest         ErrorCode = "BAD_REQUEST"
	CodeConflict           ErrorCode = "CONFLICT"
	CodeTooManyRequests    ErrorCode = "TOO_MANY_REQUESTS"
	CodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"

	// User-related errors
	CodeUserNotFound           ErrorCode = "USER_NOT_FOUND"
	CodeUserAlreadyExists      ErrorCode = "USER_ALREADY_EXISTS"
	CodeUserInactive           ErrorCode = "USER_INACTIVE"
	CodeUserLocked             ErrorCode = "USER_LOCKED"
	CodeInvalidCredentials     ErrorCode = "INVALID_CREDENTIALS"
	CodePasswordTooWeak        ErrorCode = "PASSWORD_TOO_WEAK"
	CodeInsufficientPrivileges ErrorCode = "INSUFFICIENT_PRIVILEGES"
	CodeTokenRevoked           ErrorCode = "TOKEN_REVOKED"
	CodeTokenExpired           ErrorCode = "TOKEN_EXPIRED"
	CodeTokenInvalid           ErrorCode = "TOKEN_INVALID"

	// Task-related errors
	CodeTaskNotFound      ErrorCode = "TASK_NOT_FOUND"
	CodeCategoryNotFound  ErrorCode = "CATEGORY_NOT_FOUND"
	CodeInvalidTaskStatus ErrorCode = "INVALID_TASK_STATUS"
	CodeInvalidPriority   ErrorCode = "INVALID_PRIORITY"

	// Analytics errors
	CodeInvalidArgument ErrorCode = "INVALID_ARGUMENT"
)

// AppError represents a structured application error.
type AppError struct {
	Code     ErrorCode         `json:"code"`
	Message  string            `json:"message"`
	Details  string            `json:"details,omitempty"`
	Cause    error             `json:"-"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// Unwrap returns the underlying cause.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target by code.
func (e *AppError) Is(target error) bool {
	var appErr *AppError
	if errors.As(target, &appErr) {
		return e.Code == appErr.Code
	}
	return false
}

// WithMessage returns a copy with a custom message.
func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{
		Code:     e.Code,
		Message:  msg,
		Details:  e.Details,
		Cause:    e.Cause,
		Metadata: e.Metadata,
	}
}

// WithDetails returns a copy with additional details.
func (e *AppError) WithDetails(details string) *AppError {
	return &AppError{
		Code:     e.Code,
		Message:  e.Message,
		Details:  details,
		Cause:    e.Cause,
		Metadata: e.Metadata,
	}
}

// WithCause returns a copy with the underlying cause.
func (e *AppError) WithCause(cause error) *AppError {
	return &AppError{
		Code:     e.Code,
		Message:  e.Message,
		Details:  e.Details,
		Cause:    cause,
		Metadata: e.Metadata,
	}
}

// WithMeta returns a copy with additional metadata.
func (e *AppError) WithMeta(key, value string) *AppError {
	meta := make(map[string]string, len(e.Metadata)+1)
	for k, v := range e.Metadata {
		meta[k] = v
	}
	meta[key] = value
	return &AppError{
		Code:     e.Code,
		Message:  e.Message,
		Details:  e.Details,
		Cause:    e.Cause,
		Metadata: meta,
	}
}

// HTTPStatus returns the appropriate HTTP status code.
func (e *AppError) HTTPStatus() int {
	return CodeToHTTPStatus(e.Code)
}

// GRPCCode returns the appropriate gRPC status code.
func (e *AppError) GRPCCode() codes.Code {
	return CodeToGRPCCode(e.Code)
}

// New creates a new AppError.
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with an AppError.
func Wrap(code ErrorCode, cause error, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// CodeToHTTPStatus maps error codes to HTTP status codes.
func CodeToHTTPStatus(code ErrorCode) int {
	switch code {
	case CodeValidation, CodeBadRequest, CodeInvalidArgument,
		CodeInvalidTaskStatus, CodeInvalidPriority, CodePasswordTooWeak:
		return http.StatusBadRequest

	case CodeUnauthorized, CodeInvalidCredentials, CodeTokenRevoked,
		CodeTokenExpired, CodeTokenInvalid:
		return http.StatusUnauthorized

	case CodeForbidden, CodeInsufficientPrivileges, CodeUserLocked, CodeUserInactive:
		return http.StatusForbidden

	case CodeNotFound, CodeUserNotFound, CodeTaskNotFound, CodeCategoryNotFound:
		return http.StatusNotFound

	case CodeAlreadyExists, CodeUserAlreadyExists, CodeConflict:
		return http.StatusConflict

	case CodeTooManyRequests:
		return http.StatusTooManyRequests

	case CodeServiceUnavailable:
		return http.StatusServiceUnavailable

	default:
		return http.StatusInternalServerError
	}
}

// CodeToGRPCCode maps error codes to gRPC status codes.
func CodeToGRPCCode(code ErrorCode) codes.Code {
	switch code {
	case CodeValidation, CodeBadRequest, CodeInvalidArgument,
		CodeInvalidTaskStatus, CodeInvalidPriority, CodePasswordTooWeak:
		return codes.InvalidArgument

	case CodeUnauthorized, CodeInvalidCredentials, CodeTokenRevoked,
		CodeTokenExpired, CodeTokenInvalid:
		return codes.Unauthenticated

	case CodeForbidden, CodeInsufficientPrivileges, CodeUserLocked, CodeUserInactive:
		return codes.PermissionDenied

	case CodeNotFound, CodeUserNotFound, CodeTaskNotFound, CodeCategoryNotFound:
		return codes.NotFound

	case CodeAlreadyExists, CodeUserAlreadyExists, CodeConflict:
		return codes.AlreadyExists

	case CodeTooManyRequests:
		return codes.ResourceExhausted

	case CodeServiceUnavailable:
		return codes.Unavailable

	default:
		return codes.Internal
	}
}
