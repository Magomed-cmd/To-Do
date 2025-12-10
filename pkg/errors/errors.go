package errors

import (
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
)

type ErrorCode string

const (
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

	CodeTaskNotFound      ErrorCode = "TASK_NOT_FOUND"
	CodeCategoryNotFound  ErrorCode = "CATEGORY_NOT_FOUND"
	CodeCommentNotFound   ErrorCode = "COMMENT_NOT_FOUND"
	CodeInvalidTaskStatus ErrorCode = "INVALID_TASK_STATUS"
	CodeInvalidPriority   ErrorCode = "INVALID_PRIORITY"

	CodeInvalidArgument ErrorCode = "INVALID_ARGUMENT"
)

type AppError struct {
	Code     ErrorCode         `json:"code"`
	Message  string            `json:"message"`
	Details  string            `json:"details,omitempty"`
	Cause    error             `json:"-"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func (e *AppError) Is(target error) bool {
	var appErr *AppError
	if errors.As(target, &appErr) {
		return e.Code == appErr.Code
	}
	return false
}

func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{
		Code:     e.Code,
		Message:  msg,
		Details:  e.Details,
		Cause:    e.Cause,
		Metadata: e.Metadata,
	}
}

func (e *AppError) WithDetails(details string) *AppError {
	return &AppError{
		Code:     e.Code,
		Message:  e.Message,
		Details:  details,
		Cause:    e.Cause,
		Metadata: e.Metadata,
	}
}

func (e *AppError) WithCause(cause error) *AppError {
	return &AppError{
		Code:     e.Code,
		Message:  e.Message,
		Details:  e.Details,
		Cause:    cause,
		Metadata: e.Metadata,
	}
}

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

func (e *AppError) HTTPStatus() int {
	return CodeToHTTPStatus(e.Code)
}

func (e *AppError) GRPCCode() codes.Code {
	return CodeToGRPCCode(e.Code)
}

func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func Wrap(code ErrorCode, cause error, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

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

	case CodeNotFound, CodeUserNotFound, CodeTaskNotFound, CodeCategoryNotFound, CodeCommentNotFound:
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

	case CodeNotFound, CodeUserNotFound, CodeTaskNotFound, CodeCategoryNotFound, CodeCommentNotFound:
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
