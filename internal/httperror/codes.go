package httperror

// ErrorCode represents an application error code.
type ErrorCode string

// Common error codes.
const (
	ErrCodeInvalidJSON    ErrorCode = "INVALID_JSON_FORMAT"
	ErrCodeInvalidParam   ErrorCode = "INVALID_PARAMETER"
	ErrCodeInternalServer ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED_ACCESS"
)

// Authentication and user-related error codes.
const (
	ErrCodeUserNotFound       ErrorCode = "USER_NOT_FOUND"
	ErrCodeEmailAlreadyTaken  ErrorCode = "EMAIL_ALREADY_EXISTS"
	ErrCodePasswordTooWeak    ErrorCode = "PASSWORD_TOO_WEAK"
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS" //nolint:gosec
	ErrCodeUserBlocked        ErrorCode = "USER_BLOCKED"
)
