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

var _ domain.SessionRepository = (*SessionRepository)(nil)

type SessionRepository struct {
	q *gen.Queries
}

func NewSessionRepository(q *gen.Queries) *SessionRepository {
	return &SessionRepository{q: q}
}

func (sr *SessionRepository) Save(ctx context.Context, session *domain.Session) error {
	_, err := sr.q.CreateSession(ctx, toCreateSessionParams(session))

	return err
}

func (sr *SessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	repoSession, err := sr.q.GetSessionByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("get session by id: %w", err)
	}

	return toDomainSession(&repoSession), nil
}

func (sr *SessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	repoSessions, err := sr.q.GetSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get sessions by user id: %w", err)
	}

	out := make([]*domain.Session, len(repoSessions))
	for i := range repoSessions {
		out[i] = toDomainSession(&repoSessions[i])
	}

	return out, nil
}

func (sr *SessionRepository) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	repoSession, err := sr.q.GetSessionByToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("get session by token: %w", err)
	}

	return toDomainSession(&repoSession), nil
}

func (sr *SessionRepository) Update(ctx context.Context, session *domain.Session) error {
	_, err := sr.q.UpdateSession(ctx, toUpdateSessionParams(session))

	return err
}

func (sr *SessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return sr.q.DeleteSession(ctx, id)
}

func (sr *SessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return sr.q.DeleteSessionsByUserID(ctx, userID)
}

func toCreateSessionParams(session *domain.Session) gen.CreateSessionParams {
	return gen.CreateSessionParams{
		ID:        session.ID,
		UserID:    session.UserID,
		Token:     session.Token,
		UserAgent: session.UserAgent,
		ClientIP:  session.ClientIP,
		ExpiresAt: session.ExpiresAt,
		RevokedAt: session.RevokedAt,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}
}

func toUpdateSessionParams(session *domain.Session) gen.UpdateSessionParams {
	return gen.UpdateSessionParams{
		ID:        session.ID,
		Token:     session.Token,
		UserAgent: session.UserAgent,
		ClientIP:  session.ClientIP,
		ExpiresAt: session.ExpiresAt,
		RevokedAt: session.RevokedAt,
		UpdatedAt: session.UpdatedAt,
	}
}

func toDomainSession(repoSession *gen.Session) *domain.Session {
	return &domain.Session{
		ID:        repoSession.ID,
		UserID:    repoSession.UserID,
		Token:     repoSession.Token,
		UserAgent: repoSession.UserAgent,
		ClientIP:  repoSession.ClientIP,
		ExpiresAt: repoSession.ExpiresAt,
		RevokedAt: repoSession.RevokedAt,
		CreatedAt: repoSession.CreatedAt,
		UpdatedAt: repoSession.UpdatedAt,
	}
}
