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

var _ domain.TokenRepository = (*TokenRepository)(nil)

type TokenRepository struct {
	q *gen.Queries
}

func NewTokenRepository(q *gen.Queries) *TokenRepository {
	return &TokenRepository{q: q}
}

func (tr *TokenRepository) Save(ctx context.Context, token *domain.Token) error {
	_, err := tr.q.CreateToken(ctx, toCreateTokenParams(token))

	return err
}

func (tr *TokenRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Token, error) {
	repoToken, err := tr.q.GetTokenByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("get token by id: %w", err)
	}

	return toDomainToken(&repoToken), nil
}

func (tr *TokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Token, error) {
	repoTokens, err := tr.q.GetTokensByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get tokens by user id: %w", err)
	}

	out := make([]*domain.Token, len(repoTokens))
	for i := range repoTokens {
		out[i] = toDomainToken(&repoTokens[i])
	}

	return out, nil
}

func (tr *TokenRepository) Update(ctx context.Context, token *domain.Token) error {
	_, err := tr.q.UpdateToken(ctx, toUpdateTokenParams(token))

	return err
}

func (tr *TokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return tr.q.DeleteToken(ctx, id)
}

func (tr *TokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return tr.q.DeleteTokensByUserID(ctx, userID)
}

func toCreateTokenParams(token *domain.Token) gen.CreateTokenParams {
	return gen.CreateTokenParams{
		ID:        token.ID,
		UserID:    token.UserID,
		Token:     token.Token,
		Type:      string(token.Type),
		ExpiresAt: token.ExpiresAt,
		UsedAt:    token.UsedAt,
		CreatedAt: token.CreatedAt,
	}
}

func toUpdateTokenParams(token *domain.Token) gen.UpdateTokenParams {
	return gen.UpdateTokenParams{
		ID:        token.ID,
		Token:     token.Token,
		Type:      string(token.Type),
		ExpiresAt: token.ExpiresAt,
		UsedAt:    token.UsedAt,
	}
}

func toDomainToken(repoToken *gen.Token) *domain.Token {
	return &domain.Token{
		ID:        repoToken.ID,
		UserID:    repoToken.UserID,
		Type:      domain.TokenType(repoToken.Type),
		Token:     repoToken.Token,
		ExpiresAt: repoToken.ExpiresAt,
		UsedAt:    repoToken.UsedAt,
		CreatedAt: repoToken.CreatedAt,
	}
}
