package errors

import (
	"encoding/json"
	"errors"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HTTPError struct {
	Error   string            `json:"error"`
	Code    string            `json:"code"`
	Details string            `json:"details,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
}

func WriteHTTP(w http.ResponseWriter, err error) {
	appErr := AsAppError(err)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(appErr.HTTPStatus())

	response := HTTPError{
		Error:   appErr.Message,
		Code:    string(appErr.Code),
		Details: appErr.Details,
		Meta:    appErr.Metadata,
	}

	_ = json.NewEncoder(w).Encode(response)
}

func AsAppError(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	if st, ok := status.FromError(err); ok {
		return &AppError{
			Code:    grpcCodeToErrorCode(st.Code()),
			Message: st.Message(),
			Cause:   err,
		}
	}

	return &AppError{
		Code:    CodeInternal,
		Message: "internal server error",
		Cause:   err,
	}
}

func IsCode(err error, code ErrorCode) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

func IsNotFound(err error) bool {
	return IsCode(err, CodeNotFound) ||
		IsCode(err, CodeUserNotFound) ||
		IsCode(err, CodeTaskNotFound) ||
		IsCode(err, CodeCategoryNotFound) ||
		IsCode(err, CodeCommentNotFound)
}

func IsUnauthorized(err error) bool {
	return IsCode(err, CodeUnauthorized) ||
		IsCode(err, CodeInvalidCredentials) ||
		IsCode(err, CodeTokenRevoked) ||
		IsCode(err, CodeTokenExpired) ||
		IsCode(err, CodeTokenInvalid)
}

func IsForbidden(err error) bool {
	return IsCode(err, CodeForbidden) ||
		IsCode(err, CodeInsufficientPrivileges) ||
		IsCode(err, CodeUserLocked) ||
		IsCode(err, CodeUserInactive)
}

func IsValidation(err error) bool {
	return IsCode(err, CodeValidation) ||
		IsCode(err, CodeBadRequest) ||
		IsCode(err, CodeInvalidArgument) ||
		IsCode(err, CodeInvalidTaskStatus) ||
		IsCode(err, CodeInvalidPriority) ||
		IsCode(err, CodePasswordTooWeak)
}

func grpcCodeToErrorCode(code codes.Code) ErrorCode {
	switch code {
	case codes.InvalidArgument:
		return CodeBadRequest
	case codes.Unauthenticated:
		return CodeUnauthorized
	case codes.PermissionDenied:
		return CodeForbidden
	case codes.NotFound:
		return CodeNotFound
	case codes.AlreadyExists:
		return CodeAlreadyExists
	case codes.ResourceExhausted:
		return CodeTooManyRequests
	case codes.Unavailable:
		return CodeServiceUnavailable
	default:
		return CodeInternal
	}
}
