package domain

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"regexp"
	"strings"
)

var (
	_ driver.Valuer            = Username{}
	_ sql.Scanner              = (*Username)(nil)
	_ encoding.TextMarshaler   = Username{}
	_ encoding.TextUnmarshaler = (*Username)(nil)
)

type Username struct {
	value string
}

const (
	minUsernameLength = 3
	maxUsernameLength = 20
)

var usernameRegex = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

var reservedUsernames = map[string]struct{}{
	"admin": {}, "administrator": {}, "root": {}, "null": {}, "undefined": {},
	"system": {}, "support": {}, "info": {}, "true": {}, "false": {}, "yes": {}, "no": {},
	"user": {}, "superadmin": {}, "moderator": {}, "member": {}, "guest": {},
}

func NewUsername(raw string) (Username, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return Username{}, ErrUsernameRequired
	}

	normalized := strings.ToLower(trimmed)
	if len(normalized) < minUsernameLength {
		return Username{}, ErrUsernameTooShort
	}

	if len(normalized) > maxUsernameLength {
		return Username{}, ErrUsernameTooLong
	}

	if !usernameRegex.MatchString(normalized) {
		return Username{}, ErrUsernameInvalid
	}

	if _, ok := reservedUsernames[normalized]; ok {
		return Username{}, ErrUsernameInvalid
	}

	return Username{value: normalized}, nil
}

func (u Username) String() string {
	return u.value
}

func (u Username) IsZero() bool {
	return u.value == ""
}

func (u Username) Value() (driver.Value, error) {
	if u.IsZero() {
		return nil, ErrUsernameRequired
	}

	return u.value, nil
}

func (u *Username) Scan(value any) error {
	if value == nil {
		*u = Username{}

		return nil
	}

	var str string

	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return ErrUsernameScan
	}

	username, err := NewUsername(str)
	if err != nil {
		return err
	}

	*u = username

	return nil
}

func (u Username) MarshalText() ([]byte, error) {
	return []byte(u.value), nil
}

func (u *Username) UnmarshalText(text []byte) error {
	str := string(text)

	if str == "" {
		*u = Username{}

		return nil
	}

	username, err := NewUsername(str)
	if err != nil {
		return err
	}

	*u = username

	return nil
}
