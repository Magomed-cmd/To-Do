package domain

import "todoapp/pkg/errors"

var (
	ErrUserNotFound           = errors.ErrUserNotFound
	ErrUserAlreadyExists      = errors.ErrUserAlreadyExists
	ErrInvalidCredentials     = errors.ErrInvalidCredentials
	ErrUserInactive           = errors.ErrUserInactive
	ErrUserLocked             = errors.ErrUserLocked
	ErrRefreshTokenRevoked    = errors.ErrTokenRevoked
	ErrRefreshTokenMismatch   = errors.ErrTokenInvalid.WithMessage("refresh token mismatch")
	ErrPasswordTooWeak        = errors.ErrPasswordTooWeak
	ErrInsufficientPrivileges = errors.ErrInsufficientPrivileges
	ErrDataIntegrityViolation = errors.ErrConflict.WithMessage("data integrity violation")
)
