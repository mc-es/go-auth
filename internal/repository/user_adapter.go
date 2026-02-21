package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"go-auth/internal/domain"
	"go-auth/internal/repository/gen"
)

var _ domain.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	q *gen.Queries
}

func NewUserRepository(q *gen.Queries) *UserRepository {
	return &UserRepository{q: q}
}

func (ur *UserRepository) Save(ctx context.Context, user *domain.User) error {
	_, err := ur.q.CreateUser(ctx, toCreateUserParams(user))

	return err
}

func (ur *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	repoUser, err := ur.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return toDomainUser(&repoUser)
}

func (ur *UserRepository) GetByUsername(ctx context.Context, username domain.Username) (*domain.User, error) {
	repoUser, err := ur.q.GetUserByUsername(ctx, username.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("get user by username: %w", err)
	}

	return toDomainUser(&repoUser)
}

func (ur *UserRepository) GetByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	repoUser, err := ur.q.GetUserByEmail(ctx, email.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return toDomainUser(&repoUser)
}

func (ur *UserRepository) Update(ctx context.Context, user *domain.User) error {
	_, err := ur.q.UpdateUser(ctx, toUpdateUserParams(user))

	return err
}

func (ur *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return ur.q.DeleteUser(ctx, id)
}

func (ur *UserRepository) ExistsByUsername(ctx context.Context, username domain.Username) (bool, error) {
	return ur.q.ExistsByUsername(ctx, username.String())
}

func (ur *UserRepository) ExistsByEmail(ctx context.Context, email domain.Email) (bool, error) {
	return ur.q.ExistsByEmail(ctx, email.String())
}

func toCreateUserParams(user *domain.User) gen.CreateUserParams {
	return gen.CreateUserParams{
		ID:         user.ID,
		Username:   user.Username.String(),
		Email:      user.Email.String(),
		Password:   user.Password.Hash(),
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Role:       user.Role.String(),
		Status:     user.Status.String(),
		VerifiedAt: user.VerifiedAt,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}
}

func toUpdateUserParams(user *domain.User) gen.UpdateUserParams {
	return gen.UpdateUserParams{
		ID:         user.ID,
		Username:   user.Username.String(),
		Email:      user.Email.String(),
		Password:   user.Password.Hash(),
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Role:       user.Role.String(),
		Status:     user.Status.String(),
		VerifiedAt: user.VerifiedAt,
		UpdatedAt:  user.UpdatedAt,
	}
}

func toDomainUser(repoUser *gen.User) (*domain.User, error) {
	username, err := domain.NewUsername(repoUser.Username)
	if err != nil {
		return nil, fmt.Errorf("username: %w", err)
	}

	email, err := domain.NewEmail(repoUser.Email)
	if err != nil {
		return nil, fmt.Errorf("email: %w", err)
	}

	password, err := domain.NewPasswordFromHash(repoUser.Password)
	if err != nil {
		return nil, fmt.Errorf("password: %w", err)
	}

	role, err := domain.NewRole(repoUser.Role)
	if err != nil {
		return nil, fmt.Errorf("role: %w", err)
	}

	status := domain.Status(repoUser.Status)
	if !status.IsValid() {
		return nil, fmt.Errorf("invalid status: %q", repoUser.Status)
	}

	return &domain.User{
		ID:         repoUser.ID,
		Username:   username,
		Email:      email,
		Password:   password,
		FirstName:  repoUser.FirstName,
		LastName:   repoUser.LastName,
		Role:       role,
		Status:     status,
		VerifiedAt: repoUser.VerifiedAt,
		CreatedAt:  repoUser.CreatedAt,
		UpdatedAt:  repoUser.UpdatedAt,
	}, nil
}
