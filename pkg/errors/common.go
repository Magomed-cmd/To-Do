package errors

var (
	ErrInternal           = New(CodeInternal, "internal server error")
	ErrValidation         = New(CodeValidation, "validation failed")
	ErrNotFound           = New(CodeNotFound, "resource not found")
	ErrAlreadyExists      = New(CodeAlreadyExists, "resource already exists")
	ErrForbidden          = New(CodeForbidden, "access denied")
	ErrUnauthorized       = New(CodeUnauthorized, "unauthorized")
	ErrBadRequest         = New(CodeBadRequest, "bad request")
	ErrConflict           = New(CodeConflict, "conflict detected")
	ErrTooManyRequests    = New(CodeTooManyRequests, "too many requests")
	ErrServiceUnavailable = New(CodeServiceUnavailable, "service temporarily unavailable")
	ErrInvalidArgument    = New(CodeInvalidArgument, "invalid argument")
)

var (
	ErrUserNotFound           = New(CodeUserNotFound, "user not found")
	ErrUserAlreadyExists      = New(CodeUserAlreadyExists, "user already exists")
	ErrUserInactive           = New(CodeUserInactive, "user is inactive")
	ErrUserLocked             = New(CodeUserLocked, "user account is locked")
	ErrInvalidCredentials     = New(CodeInvalidCredentials, "invalid credentials")
	ErrPasswordTooWeak        = New(CodePasswordTooWeak, "password does not meet requirements")
	ErrInsufficientPrivileges = New(CodeInsufficientPrivileges, "insufficient privileges")
	ErrTokenRevoked           = New(CodeTokenRevoked, "token has been revoked")
	ErrTokenExpired           = New(CodeTokenExpired, "token has expired")
	ErrTokenInvalid           = New(CodeTokenInvalid, "invalid token")
)

var (
	ErrTaskNotFound      = New(CodeTaskNotFound, "task not found")
	ErrCategoryNotFound  = New(CodeCategoryNotFound, "category not found")
	ErrCommentNotFound   = New(CodeCommentNotFound, "comment not found")
	ErrInvalidTaskStatus = New(CodeInvalidTaskStatus, "invalid task status")
	ErrInvalidPriority   = New(CodeInvalidPriority, "invalid priority")
)
