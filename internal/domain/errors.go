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
