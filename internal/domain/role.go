package domain

import (
	"database/sql/driver"
	"strings"
)

type Permission string

const (
	PermUserRead   Permission = "user:read"
	PermUserWrite  Permission = "user:write"
	PermUserBan    Permission = "user:ban"
	PermUserDelete Permission = "user:delete"
)

type Role struct {
	value string
}

const (
	RoleUser       = "user"
	RoleAdmin      = "admin"
	RoleSuperAdmin = "superadmin"
)

var rolePermissions = map[string]map[Permission]struct{}{
	RoleUser: {
		PermUserRead: {},
	},
	RoleAdmin: {
		PermUserRead:  {},
		PermUserWrite: {},
		PermUserBan:   {},
	},
	RoleSuperAdmin: {
		PermUserRead:   {},
		PermUserWrite:  {},
		PermUserBan:    {},
		PermUserDelete: {},
	},
}

func NewRole(raw string) (Role, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return Role{}, ErrRoleRequired
	}

	normalized := strings.ToLower(trimmed)
	if _, ok := rolePermissions[normalized]; !ok {
		return Role{}, ErrRoleInvalid
	}

	return Role{value: normalized}, nil
}

func (r Role) String() string {
	return r.value
}

func (r Role) IsZero() bool {
	return r.value == ""
}

func (r Role) HasPermission(perm Permission) bool {
	_, ok := rolePermissions[r.value][perm]

	return ok
}

func (r Role) Value() (driver.Value, error) {
	if r.IsZero() {
		return nil, ErrRoleRequired
	}

	return r.value, nil
}

func (r *Role) Scan(value any) error {
	if value == nil {
		*r = Role{}

		return nil
	}

	var str string

	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return ErrRoleScan
	}

	role, err := NewRole(str)
	if err != nil {
		return err
	}

	*r = role

	return nil
}

func (r Role) MarshalText() ([]byte, error) {
	return []byte(r.value), nil
}

func (r *Role) UnmarshalText(text []byte) error {
	str := string(text)

	if str == "" {
		*r = Role{}

		return nil
	}

	role, err := NewRole(str)
	if err != nil {
		return err
	}

	*r = role

	return nil
}
