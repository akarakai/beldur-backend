package account

import (
	"beldur/internal/domain/account/hash"
	"context"
	"fmt"
)

type Service struct {
	accountRepo Repository
}

func NewService(repo Repository) *Service {
	if repo == nil {
		panic("account: nil Repository")
	}

	return &Service{
		accountRepo: repo,
	}
}

func (s *Service) CreateAccount(ctx context.Context, request CreateAccountRequest) (*Account, error) {
	newAcc, err := s.validateAndBuildNewAccount(request)
	if err != nil {
		return nil, err
	}

	accAfterSave, err := s.accountRepo.Save(ctx, newAcc)
	if err != nil {
		return nil, err
	}

	return accAfterSave, nil
}

func (s *Service) validateAndBuildNewAccount(request CreateAccountRequest) (*Account, error) {
	pwdHash, err := hash.HashPassword(request.Password)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrHashing, err)
	}

	if request.Email == "" {
		return New(request.Username, pwdHash)
	}

	return New(
		request.Username,
		pwdHash,
		WithEmail(request.Email),
	)
}
