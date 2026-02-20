package domain

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Save(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByUsername(ctx context.Context, username Username) (*User, error)
	GetByEmail(ctx context.Context, email Email) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsByUsername(ctx context.Context, username Username) (bool, error)
	ExistsByEmail(ctx context.Context, email Email) (bool, error)
}

type SessionRepository interface {
	Save(ctx context.Context, session *Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*Session, error)
	GetByToken(ctx context.Context, token string) (*Session, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Session, error)
	Update(ctx context.Context, session *Session) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}
