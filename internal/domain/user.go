package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID
	Username   Username
	Email      Email
	Password   Password
	FirstName  string
	LastName   string
	Role       Role
	Status     Status
	VerifiedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewUser(username Username, email Email, password Password, firstName, lastName string) (*User, error) {
	if username.IsZero() {
		return nil, ErrUsernameRequired
	}

	if email.IsZero() {
		return nil, ErrEmailRequired
	}

	if password.IsZero() {
		return nil, ErrPasswordRequired
	}

	fName := strings.TrimSpace(firstName)
	lName := strings.TrimSpace(lastName)

	if fName == "" {
		return nil, ErrFirstNameRequired
	}

	if lName == "" {
		return nil, ErrLastNameRequired
	}

	role, err := NewRole(RoleUser)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	return &User{
		ID:        uuid.New(),
		Username:  username,
		Email:     email,
		Password:  password,
		FirstName: fName,
		LastName:  lName,
		Role:      role,
		Status:    StatusActivated,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

func (u *User) IsActivated() bool {
	return u.Status == StatusActivated
}

func (u *User) IsBanned() bool {
	return u.Status == StatusBanned
}

func (u *User) IsVerified() bool {
	return u.VerifiedAt != nil
}

func (u *User) CanLogin() bool {
	return u.IsActivated() && u.IsVerified()
}

func (u *User) Verify() error {
	if !u.IsActivated() {
		return ErrUserNotActivated
	}

	if u.IsVerified() {
		return ErrUserVerified
	}

	now := time.Now().UTC()
	u.VerifiedAt = &now
	u.touch()

	return nil
}

func (u *User) Ban() error {
	if u.IsBanned() {
		return ErrUserBanned
	}

	u.Status = StatusBanned
	u.touch()

	return nil
}

func (u *User) Unban() error {
	if !u.IsBanned() {
		return ErrUserNotBanned
	}

	u.Status = StatusActivated
	u.touch()

	return nil
}

func (u *User) ChangePassword(newPassword Password) error {
	if !u.IsActivated() {
		return ErrUserNotActivated
	}

	u.Password = newPassword
	u.touch()

	return nil
}

func (u *User) UpdateInfo(firstName, lastName string) error {
	if !u.IsActivated() {
		return ErrUserNotActivated
	}

	fName := strings.TrimSpace(firstName)
	lName := strings.TrimSpace(lastName)

	if fName == "" {
		return ErrFirstNameRequired
	}

	if lName == "" {
		return ErrLastNameRequired
	}

	u.FirstName = fName
	u.LastName = lName
	u.touch()

	return nil
}

func (u *User) UpdateRole(role Role) error {
	if !u.IsActivated() {
		return ErrUserNotActivated
	}

	u.Role = role
	u.touch()

	return nil
}

func (u *User) touch() {
	u.UpdatedAt = time.Now().UTC()
}
