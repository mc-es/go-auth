package apperror

type Code string

// Common error codes.
const (
	ErrCodeInvalidJSON    Code = "INVALID_JSON_FORMAT"
	ErrCodeInvalidParam   Code = "INVALID_PARAMETER"
	ErrCodeInternalServer Code = "INTERNAL_SERVER_ERROR"
	ErrCodeUnauthorized   Code = "UNAUTHORIZED_ACCESS"
)

// User error codes.
const (
	ErrCodeUserNotFound       Code = "USER_NOT_FOUND"
	ErrCodeEmailAlreadyUsed   Code = "EMAIL_ALREADY_USED"
	ErrCodePasswordTooWeak    Code = "PASSWORD_TOO_WEAK"
	ErrCodeInvalidCredentials Code = "INVALID_CREDENTIALS" //nolint:gosec
	ErrCodeUserBlocked        Code = "USER_BLOCKED"
)
