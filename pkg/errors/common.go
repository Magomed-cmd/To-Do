package errors

// Predefined application errors for common scenarios.
// Import this package and use these errors directly or customize with With* methods.

// === Generic Errors ===

var (
	// ErrInternal represents an unexpected internal error.
	ErrInternal = New(CodeInternal, "internal server error")

	// ErrValidation represents a validation failure.
	ErrValidation = New(CodeValidation, "validation failed")

	// ErrNotFound represents a resource not found.
	ErrNotFound = New(CodeNotFound, "resource not found")

	// ErrAlreadyExists represents a resource conflict.
	ErrAlreadyExists = New(CodeAlreadyExists, "resource already exists")

	// ErrForbidden represents an access denied error.
	ErrForbidden = New(CodeForbidden, "access denied")

	// ErrUnauthorized represents an authentication failure.
	ErrUnauthorized = New(CodeUnauthorized, "unauthorized")

	// ErrBadRequest represents a malformed request.
	ErrBadRequest = New(CodeBadRequest, "bad request")

	// ErrConflict represents a data conflict.
	ErrConflict = New(CodeConflict, "conflict detected")

	// ErrTooManyRequests represents rate limiting.
	ErrTooManyRequests = New(CodeTooManyRequests, "too many requests")

	// ErrServiceUnavailable represents a service outage.
	ErrServiceUnavailable = New(CodeServiceUnavailable, "service temporarily unavailable")

	// ErrInvalidArgument represents an invalid input argument.
	ErrInvalidArgument = New(CodeInvalidArgument, "invalid argument")
)

// === User Errors ===

var (
	// ErrUserNotFound is returned when user lookup fails.
	ErrUserNotFound = New(CodeUserNotFound, "user not found")

	// ErrUserAlreadyExists is returned on duplicate registration.
	ErrUserAlreadyExists = New(CodeUserAlreadyExists, "user already exists")

	// ErrUserInactive is returned when an inactive user tries to act.
	ErrUserInactive = New(CodeUserInactive, "user is inactive")

	// ErrUserLocked is returned when account is locked.
	ErrUserLocked = New(CodeUserLocked, "user account is locked")

	// ErrInvalidCredentials is returned on auth failure.
	ErrInvalidCredentials = New(CodeInvalidCredentials, "invalid credentials")

	// ErrPasswordTooWeak is returned when password doesn't meet policy.
	ErrPasswordTooWeak = New(CodePasswordTooWeak, "password does not meet requirements")

	// ErrInsufficientPrivileges is returned on permission check failure.
	ErrInsufficientPrivileges = New(CodeInsufficientPrivileges, "insufficient privileges")

	// ErrTokenRevoked is returned when refresh token is revoked.
	ErrTokenRevoked = New(CodeTokenRevoked, "token has been revoked")

	// ErrTokenExpired is returned when token is expired.
	ErrTokenExpired = New(CodeTokenExpired, "token has expired")

	// ErrTokenInvalid is returned when token is malformed.
	ErrTokenInvalid = New(CodeTokenInvalid, "invalid token")
)

// === Task Errors ===

var (
	// ErrTaskNotFound is returned when task doesn't exist or user has no access.
	ErrTaskNotFound = New(CodeTaskNotFound, "task not found")

	// ErrCategoryNotFound is returned when category lookup fails.
	ErrCategoryNotFound = New(CodeCategoryNotFound, "category not found")

	// ErrInvalidTaskStatus is returned on invalid status transition.
	ErrInvalidTaskStatus = New(CodeInvalidTaskStatus, "invalid task status")

	// ErrInvalidPriority is returned on unsupported priority.
	ErrInvalidPriority = New(CodeInvalidPriority, "invalid priority")
)
