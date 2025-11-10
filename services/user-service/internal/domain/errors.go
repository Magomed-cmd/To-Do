package domain

import (
	"fmt"
)

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
	// Основные пользовательские ошибки
	ErrUserNotFound       = &DomainError{Code: "USER_NOT_FOUND", Message: "user not found", When: "returned when lookup by identifier yields no user"}
	ErrUserAlreadyExists  = &DomainError{Code: "USER_ALREADY_EXISTS", Message: "user already exists", When: "returned when creating a user with an email or github id that is already registered"}
	ErrInvalidCredentials = &DomainError{Code: "INVALID_CREDENTIALS", Message: "invalid credentials", When: "returned when password or login factors do not match stored credentials"}
	ErrUserInactive       = &DomainError{Code: "USER_INACTIVE", Message: "user is inactive", When: "returned when an inactive user attempts to authenticate or perform restricted operations"}
	ErrUserLocked         = &DomainError{Code: "USER_LOCKED", Message: "user is locked", When: "returned when account is locked due to security policies"}
	ErrUserSuspended      = &DomainError{Code: "USER_SUSPENDED", Message: "user is suspended", When: "returned when suspended account tries to access the system"}

	// JWT и токены
	ErrRefreshTokenRevoked  = &DomainError{Code: "REFRESH_TOKEN_REVOKED", Message: "refresh token revoked", When: "returned when refresh token is revoked during rotation or logout"}
	ErrRefreshTokenMismatch = &DomainError{Code: "REFRESH_TOKEN_MISMATCH", Message: "refresh token mismatch", When: "returned when presented refresh token does not match stored hash"}

	// Пароли
	ErrPasswordTooWeak = &DomainError{Code: "PASSWORD_TOO_WEAK", Message: "password does not meet requirements", When: "returned when password update fails strength policy"}

	// Роли и права доступа
	ErrRoleTransitionNotAllowed = &DomainError{Code: "ROLE_TRANSITION_NOT_ALLOWED", Message: "role transition not allowed", When: "returned when requested role change violates promotion rules"}
	ErrInsufficientPrivileges   = &DomainError{Code: "INSUFFICIENT_PRIVILEGES", Message: "insufficient privileges", When: "returned when user lacks required permission for action"}

	// Настройки пользователя
	ErrPreferenceNotFound = &DomainError{Code: "PREFERENCE_NOT_FOUND", Message: "preference not found", When: "returned when requested user preference key is absent"}
	ErrPreferenceConflict = &DomainError{Code: "PREFERENCE_CONFLICT", Message: "preference conflict detected", When: "returned when preference update violates constraints or conflicts with state"}

	// Уведомления
	ErrNotificationOptedOut = &DomainError{Code: "NOTIFICATION_OPTED_OUT", Message: "notification opted out", When: "returned when user opted out of the notification type"}

	// Общие системные ошибки
	ErrDataIntegrityViolation = &DomainError{Code: "DATA_INTEGRITY_VIOLATION", Message: "data integrity violation detected", When: "returned when database constraints fail or inconsistent state detected"}
	ErrInvariantViolation     = &DomainError{Code: "INVARIANT_VIOLATION", Message: "domain invariant violation", When: "returned when business rule invariants are broken"}
)
