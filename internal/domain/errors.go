package domain

import "errors"

var (
	ErrEmailRequired = errors.New("domain: email is required")
	ErrEmailInvalid  = errors.New("domain: email is invalid")
	ErrEmailScan     = errors.New("domain: unsupported type for email")
)

var (
	ErrPasswordRequired = errors.New("domain: password is required")
	ErrPasswordScan     = errors.New("domain: unsupported type for password")
)

var (
	ErrRoleRequired = errors.New("domain: role is required")
	ErrRoleInvalid  = errors.New("domain: role is invalid")
	ErrRoleScan     = errors.New("domain: unsupported type for role")
)

var (
	ErrUsernameRequired = errors.New("domain: username is required")
	ErrUsernameTooShort = errors.New("domain: username is too short")
	ErrUsernameTooLong  = errors.New("domain: username is too long")
	ErrUsernameInvalid  = errors.New("domain: username is invalid")
	ErrUsernameScan     = errors.New("domain: unsupported type for username")
)

var (
	ErrFirstNameRequired = errors.New("domain: first name is required")
	ErrLastNameRequired  = errors.New("domain: last name is required")
	ErrUserBanned        = errors.New("domain: user is banned")
	ErrUserNotBanned     = errors.New("domain: user is not banned")
	ErrUserNotActivated  = errors.New("domain: user is not activated")
	ErrUserVerified      = errors.New("domain: user is verified")
)

var (
	ErrSessionExpired          = errors.New("domain: session is expired")
	ErrSessionRevoked          = errors.New("domain: session is revoked")
	ErrUserIDRequired          = errors.New("domain: user ID is required")
	ErrTokenRequired           = errors.New("domain: token is required")
	ErrTokenInvalid            = errors.New("domain: token is invalid")
	ErrTokenExpired            = errors.New("domain: token is expired")
	ErrTokenTypeInvalid        = errors.New("domain: token type is invalid")
	ErrTokenSecretRequired     = errors.New("domain: token secret is required")
	ErrTokenAccessTTLRequired  = errors.New("domain: access TTL must be positive")
	ErrTokenRefreshTTLRequired = errors.New("domain: refresh TTL must be positive")
)
