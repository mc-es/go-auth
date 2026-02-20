package domain

import (
	"database/sql/driver"
)

type Password struct {
	value string
}

func NewPasswordFromHash(hash string) (Password, error) {
	if hash == "" {
		return Password{}, ErrPasswordRequired
	}

	return Password{value: hash}, nil
}

func (p Password) String() string {
	return "*****"
}

func (p Password) IsZero() bool {
	return p.value == ""
}

func (p Password) Hash() string {
	return p.value
}

func (p Password) Value() (driver.Value, error) {
	if p.IsZero() {
		return nil, ErrPasswordRequired
	}

	return p.value, nil
}

func (p *Password) Scan(value any) error {
	if value == nil {
		*p = Password{}

		return nil
	}

	var str string

	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return ErrPasswordScan
	}

	password, err := NewPasswordFromHash(str)
	if err != nil {
		return err
	}

	*p = password

	return nil
}

func (p Password) MarshalText() ([]byte, error) {
	return []byte("*****"), nil
}
