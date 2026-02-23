package service

import (
	"context"

	"go-auth/internal/apperror"
	"go-auth/internal/domain"
)

func (s *service) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	username, email, err := s.validateRequest(req)
	if err != nil {
		return nil, err
	}

	if checkErr := s.checkConflicts(ctx, username, email); checkErr != nil {
		return nil, checkErr
	}

	password, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgOperationFailed, err)
	}

	user, err := domain.NewUser(username, email, password, req.FirstName, req.LastName)
	if err != nil {
		return nil, apperror.BadRequest(
			apperror.ErrCodeInvalidParam,
			err.Error(),
			err,
		)
	}

	if saveErr := s.userRepo.Save(ctx, user); saveErr != nil {
		return nil, apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgOperationFailed, saveErr)
	}

	return &RegisterResponse{
		UserID: user.ID,
	}, nil
}

func (s *service) validateRequest(req *RegisterRequest) (domain.Username, domain.Email, error) {
	if req == nil {
		return domain.Username{}, domain.Email{}, apperror.BadRequest(
			apperror.ErrCodeInvalidParam,
			apperror.MsgRegisterRequestRequired,
			nil,
		)
	}

	username, err := domain.NewUsername(req.Username)
	if err != nil {
		return domain.Username{}, domain.Email{}, apperror.BadRequest(
			apperror.ErrCodeInvalidParam,
			err.Error(),
			err,
		)
	}

	email, err := domain.NewEmail(req.Email)
	if err != nil {
		return domain.Username{}, domain.Email{}, apperror.BadRequest(
			apperror.ErrCodeInvalidParam,
			err.Error(),
			err,
		)
	}

	return username, email, nil
}

func (s *service) checkConflicts(ctx context.Context, username domain.Username, email domain.Email) error {
	exists, err := s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgOperationFailed, err)
	}

	if exists {
		return apperror.Conflict(apperror.ErrCodeUsernameAlreadyUsed, apperror.MsgUsernameAlreadyInUse, nil)
	}

	exists, err = s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return apperror.InternalServerError(apperror.ErrCodeInternalServer, apperror.MsgOperationFailed, err)
	}

	if exists {
		return apperror.Conflict(apperror.ErrCodeEmailAlreadyUsed, apperror.MsgEmailAlreadyInUse, nil)
	}

	return nil
}
