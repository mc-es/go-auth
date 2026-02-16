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
