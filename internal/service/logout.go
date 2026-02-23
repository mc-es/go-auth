package service

import (
	"context"

	"go-auth/internal/apperror"
)

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return apperror.BadRequest(apperror.ErrCodeInvalidParam, apperror.MsgRefreshTokenRequired, nil)
	}

	refreshTokenHash, err := s.opaqueTokenManager.Hash(refreshToken)
	if err != nil {
		return apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgOperationFailed, err)
	}

	session, err := s.sessionRepo.GetByToken(ctx, refreshTokenHash)
	if err != nil {
		return apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgGetSession, err)
	}

	if session == nil {
		return apperror.NotFound(apperror.ErrCodeSessionNotFound, apperror.MsgSessionNotFound, nil)
	}

	if session.IsExpired() || session.IsRevoked() {
		return apperror.Unauthorized(apperror.ErrCodeInvalidToken, apperror.MsgSessionExpiredOrRevoked, nil)
	}

	if err := session.Revoke(); err != nil {
		return apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgRevokeSession, err)
	}

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgUpdateSession, err)
	}

	return nil
}
