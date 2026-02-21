package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"go-auth/internal/domain"
	"go-auth/internal/repository/gen"
)

func NewRepositories(pool *pgxpool.Pool) (domain.UserRepository, domain.SessionRepository) {
	q := gen.New(pool)

	return NewUserRepository(q), NewSessionRepository(q)
}
