package domain

import "fmt"

type DomainError struct {
	Code    string
	Message string
	When    string
}

func (e *DomainError) Error() string {
	return e.Message
}

func (e *DomainError) WithMessage(message string) *DomainError {
	return &DomainError{Code: e.Code, Message: message, When: e.When}
}

func (e *DomainError) WithDetail(detail string) *DomainError {
	return &DomainError{Code: e.Code, Message: fmt.Sprintf("%s: %s", e.Message, detail), When: e.When}
}

var (
	ErrTaskNotFound         = &DomainError{Code: "TASK_NOT_FOUND", Message: "task not found", When: "returned when requested task does not exist or user has no access"}
	ErrCategoryNotFound     = &DomainError{Code: "CATEGORY_NOT_FOUND", Message: "category not found", When: "returned when category lookup fails for provided identifier"}
	ErrCommentNotFound      = &DomainError{Code: "COMMENT_NOT_FOUND", Message: "comment not found", When: "returned when task comment does not belong to the user"}
	ErrUnknownUser          = &DomainError{Code: "USER_NOT_FOUND", Message: "user not found", When: "returned when associated user record cannot be resolved via user-service"}
	ErrInvalidTaskStatus    = &DomainError{Code: "INVALID_TASK_STATUS", Message: "invalid task status", When: "returned when status transition violates workflow"}
	ErrInvalidTaskPriority  = &DomainError{Code: "INVALID_TASK_PRIORITY", Message: "invalid task priority", When: "returned when provided priority is not supported"}
	ErrForbiddenTaskAccess  = &DomainError{Code: "FORBIDDEN_TASK_ACCESS", Message: "task access denied", When: "returned when user tries to modify someone else's task"}
	ErrInvalidRecurringRule = &DomainError{Code: "INVALID_RECURRING_RULE", Message: "invalid recurrence rule", When: "returned when recurring tasks have malformed configuration"}
	ErrValidationFailed     = &DomainError{Code: "VALIDATION_FAILED", Message: "validation failed", When: "returned when business validation errors are detected"}
)
