package domain

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"net/mail"
	"strings"
)

var (
	_ driver.Valuer            = Email{}
	_ sql.Scanner              = (*Email)(nil)
	_ encoding.TextMarshaler   = Email{}
	_ encoding.TextUnmarshaler = (*Email)(nil)
)

type Email struct {
	value string
}

func NewEmail(raw string) (Email, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return Email{}, ErrEmailRequired
	}

	addr, err := mail.ParseAddress(trimmed)
	if err != nil {
		return Email{}, ErrEmailInvalid
	}

	return Email{value: strings.ToLower(addr.Address)}, nil
}

func (e Email) String() string {
	return e.value
}

func (e Email) IsZero() bool {
	return e.value == ""
}

func (e Email) Local() string {
	local, _, _ := strings.Cut(e.value, "@")

	return local
}

func (e Email) Domain() string {
	_, domain, _ := strings.Cut(e.value, "@")

	return domain
}

func (e Email) Mask() string {
	if e.IsZero() {
		return ""
	}

	local, domain, _ := strings.Cut(e.value, "@")

	runes := []rune(local)

	if len(runes) <= 2 {
		return string(runes) + "***@" + domain
	}

	return string(runes[0]) + "***" + string(runes[len(runes)-1]) + "@" + domain
}

func (e Email) Value() (driver.Value, error) {
	if e.IsZero() {
		return nil, ErrEmailRequired
	}

	return e.value, nil
}

func (e *Email) Scan(value any) error {
	if value == nil {
		*e = Email{}

		return nil
	}

	var str string

	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return ErrEmailScan
	}

	email, err := NewEmail(str)
	if err != nil {
		return err
	}

	*e = email

	return nil
}

func (e Email) MarshalText() ([]byte, error) {
	return []byte(e.value), nil
}

func (e *Email) UnmarshalText(text []byte) error {
	str := string(text)

	if str == "" {
		*e = Email{}

		return nil
	}

	email, err := NewEmail(str)
	if err != nil {
		return err
	}

	*e = email

	return nil
}
