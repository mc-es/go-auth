package domain

import "errors"

var (
	ErrEmailRequired = errors.New("email is required")
	ErrEmailInvalid  = errors.New("email is invalid")
	ErrEmailScan     = errors.New("unsupported type for email")
)

var (
	ErrPasswordRequired = errors.New("password is required")
	ErrPasswordScan     = errors.New("unsupported type for password")
)

var (
	ErrRoleRequired = errors.New("role is required")
	ErrRoleInvalid  = errors.New("role is invalid")
	ErrRoleScan     = errors.New("unsupported type for role")
)

var (
	ErrUsernameRequired = errors.New("username is required")
	ErrUsernameTooShort = errors.New("username is too short")
	ErrUsernameTooLong  = errors.New("username is too long")
	ErrUsernameInvalid  = errors.New("username is invalid")
	ErrUsernameScan     = errors.New("unsupported type for username")
)

var (
	ErrFirstNameRequired = errors.New("first name is required")
	ErrLastNameRequired  = errors.New("last name is required")
	ErrUserBanned        = errors.New("user is banned")
	ErrUserNotBanned     = errors.New("user is not banned")
	ErrUserNotActivated  = errors.New("user is not activated")
	ErrUserVerified      = errors.New("user is verified")
)

var (
	ErrSessionExpired          = errors.New("session is expired")
	ErrSessionRevoked          = errors.New("session is revoked")
	ErrUserIDRequired          = errors.New("user ID is required")
	ErrTokenRequired           = errors.New("token is required")
	ErrTokenInvalid            = errors.New("token is invalid")
	ErrTokenExpired            = errors.New("token is expired")
	ErrTokenUsed               = errors.New("token is used")
	ErrTokenTypeInvalid        = errors.New("token type is invalid")
	ErrTokenSecretRequired     = errors.New("token secret is required")
	ErrTokenAccessTTLRequired  = errors.New("access TTL must be positive")
	ErrTokenRefreshTTLRequired = errors.New("refresh TTL must be positive")
)
