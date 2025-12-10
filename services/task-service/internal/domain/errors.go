package domain

import "todoapp/pkg/errors"

var (
	ErrTaskNotFound        = errors.ErrTaskNotFound
	ErrCategoryNotFound    = errors.ErrCategoryNotFound
	ErrCommentNotFound     = errors.ErrCommentNotFound
	ErrUnknownUser         = errors.ErrUserNotFound
	ErrInvalidTaskStatus   = errors.ErrInvalidTaskStatus
	ErrInvalidTaskPriority = errors.ErrInvalidPriority
	ErrForbiddenTaskAccess = errors.ErrForbidden.WithMessage("task access denied")
	ErrValidationFailed    = errors.ErrValidation
)
