package service

import (
	"context"
	"time"

	"go-auth/internal/apperror"
	"go-auth/internal/domain"
)

func (s *service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	if req == nil {
		return nil, apperror.BadRequest(apperror.ErrCodeInvalidParam, apperror.MsgLoginRequestRequired, nil)
	}

	user, err := s.resolveUserByLogin(ctx, req.Login)
	if err != nil {
		return nil, err
	}

	if !s.passwordHasher.Compare(req.Password, user.Password) {
		return nil, apperror.Unauthorized(apperror.ErrCodeInvalidCredentials, apperror.MsgInvalidCredentials, nil)
	}

	if user.IsBanned() || !user.CanLogin() {
		return nil, apperror.Forbidden(apperror.ErrCodeUserBlocked, apperror.MsgAccountAccessRevoked, nil)
	}

	return s.createSession(ctx, user, req)
}

func (s *service) createSession(ctx context.Context, user *domain.User, req *LoginRequest) (*LoginResponse, error) {
	refreshToken, err := s.opaqueTokenManager.Generate()
	if err != nil {
		return nil, apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgOperationFailed, err)
	}

	refreshTokenHash, err := s.opaqueTokenManager.Hash(refreshToken)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgOperationFailed, err)
	}

	now := time.Now().UTC()
	refreshExpiresAt := now.Add(s.refreshTokenTTL)

	session, err := domain.NewSession(user.ID, refreshTokenHash, req.UserAgent, req.ClientIP, refreshExpiresAt)
	if err != nil {
		return nil, apperror.BadRequest(apperror.ErrCodeInvalidParam, err.Error(), err)
	}

	if saveErr := s.sessionRepo.Save(ctx, session); saveErr != nil {
		return nil, apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgOperationFailed, saveErr)
	}

	accessExpiresAt := now.Add(s.accessTokenTTL)

	accessToken, err := s.accessTokenManager.Generate(domain.AccessClaims{
		UserID: user.ID,
		Role:   user.Role,
	})
	if err != nil {
		return nil, apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgOperationFailed, err)
	}

	return &LoginResponse{
		UserID:           user.ID,
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

func (s *service) resolveUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	if u, err := domain.NewUsername(login); err == nil {
		user, err := s.userRepo.GetByUsername(ctx, u)
		if err != nil {
			return nil, apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgGetUserByUsername, err)
		}

		if user != nil {
			return user, nil
		}
	}

	if e, err := domain.NewEmail(login); err == nil {
		user, err := s.userRepo.GetByEmail(ctx, e)
		if err != nil {
			return nil, apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgGetUserByEmail, err)
		}

		if user != nil {
			return user, nil
		}
	}

	return nil, apperror.Unauthorized(apperror.ErrCodeInvalidCredentials, apperror.MsgInvalidCredentials, nil)
}
