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
	ErrCodeUserNotFound        Code = "USER_NOT_FOUND"
	ErrCodeUsernameAlreadyUsed Code = "USERNAME_ALREADY_USED"
	ErrCodeEmailAlreadyUsed    Code = "EMAIL_ALREADY_USED"
	ErrCodePasswordTooWeak     Code = "PASSWORD_TOO_WEAK"
	ErrCodeInvalidCredentials  Code = "INVALID_CREDENTIALS" //nolint:gosec
	ErrCodeUserBlocked         Code = "USER_BLOCKED"
	ErrCodeSessionNotFound     Code = "SESSION_NOT_FOUND"
	ErrCodeInvalidToken        Code = "INVALID_TOKEN"
	ErrCodeTokenRequired       Code = "TOKEN_REQUIRED"
)
