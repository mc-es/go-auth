package service

import (
	"context"
	"time"

	"go-auth/internal/apperror"
	"go-auth/internal/domain"
)

func (s *service) Refresh(ctx context.Context, req *RefreshRequest) (*RefreshResponse, error) {
	if req == nil {
		return nil, apperror.BadRequest(apperror.ErrCodeInvalidParam, apperror.MsgRefreshRequestRequired, nil)
	}

	if req.RefreshToken == "" {
		return nil, apperror.BadRequest(apperror.ErrCodeTokenRequired, apperror.MsgRefreshTokenRequired, nil)
	}

	refreshTokenHash, err := s.opaqueTokenManager.Hash(req.RefreshToken)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgOperationFailed, err)
	}

	session, err := s.getSessionForRefresh(ctx, refreshTokenHash)
	if err != nil {
		return nil, err
	}

	newSession, newRefreshToken, err := s.rotateSession(ctx, session, req.UserAgent, req.ClientIP)
	if err != nil {
		return nil, err
	}

	return s.buildRefresh(ctx, newSession, newRefreshToken)
}

func (s *service) getSessionForRefresh(ctx context.Context, refreshTokenHash string) (*domain.Session, error) {
	session, err := s.sessionRepo.GetByToken(ctx, refreshTokenHash)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgGetSession, err)
	}

	if session == nil {
		return nil, apperror.NotFound(apperror.ErrCodeSessionNotFound, apperror.MsgSessionNotFound, nil)
	}

	if !session.IsActive() {
		return nil, apperror.Unauthorized(apperror.ErrCodeInvalidToken, apperror.MsgSessionNotActive, nil)
	}

	return session, nil
}

func (s *service) rotateSession(
	ctx context.Context,
	session *domain.Session,
	userAgent, clientIP string,
) (*domain.Session, string, error) {
	newRefreshToken, err := s.opaqueTokenManager.Generate()
	if err != nil {
		return nil, "", apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgGenerateRefreshToken, err)
	}

	newRefreshTokenHash, err := s.opaqueTokenManager.Hash(newRefreshToken)
	if err != nil {
		return nil, "", apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgHashRefreshToken, err)
	}

	now := time.Now().UTC()
	newRefreshExpiresAt := now.Add(s.refreshTokenTTL)

	newSession, err := session.Rotate(newRefreshTokenHash, newRefreshExpiresAt, userAgent, clientIP)
	if err != nil {
		return nil, "", apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgRotateSession, err)
	}

	if err = s.sessionRepo.Update(ctx, session); err != nil {
		return nil, "", apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgUpdateSession, err)
	}

	if err = s.sessionRepo.Save(ctx, newSession); err != nil {
		return nil, "", apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgSaveNewSession, err)
	}

	return newSession, newRefreshToken, nil
}

func (s *service) buildRefresh(ctx context.Context, newSes *domain.Session, newTkn string) (*RefreshResponse, error) {
	user, err := s.userRepo.GetByID(ctx, newSes.UserID)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgGetUser, err)
	}

	if user == nil {
		return nil, apperror.NotFound(apperror.ErrCodeUserNotFound, apperror.MsgUserNotFound, nil)
	}

	if user.IsBanned() || !user.CanLogin() {
		return nil, apperror.Forbidden(apperror.ErrCodeUserBlocked, apperror.MsgAccountAccessRevoked, nil)
	}

	now := time.Now().UTC()
	accessExpiresAt := now.Add(s.accessTokenTTL)

	accessToken, err := s.accessTokenManager.Generate(domain.AccessClaims{
		UserID: user.ID,
		Role:   user.Role,
	})
	if err != nil {
		return nil, apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgGenerateAccessToken, err)
	}

	return &RefreshResponse{
		AccessToken:      accessToken,
		RefreshToken:     newTkn,
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: newSes.ExpiresAt,
	}, nil
}
