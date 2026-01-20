package usecase

import (
	"beldur/internal/db/tx"
	"beldur/internal/domain/account"
	"beldur/internal/domain/player"
	"context"
	"log/slog"
)

type RegisterAccount struct {
	// when a struct uses this interface, it means that it will use a transaction
	tx             tx.Transactor
	accountService *account.Service
	playerService  *player.Service
}

func NewRegisterAccount(tx tx.Transactor, accountService *account.Service, playerService *player.Service) *RegisterAccount {
	return &RegisterAccount{
		tx:             tx,
		accountService: accountService,
		playerService:  playerService,
	}
}

func (r *RegisterAccount) Execute(ctx context.Context, request CreateAccountRequest) (CreateAccountResponse, error) {
	accReq := account.CreateAccountRequest{
		Username: request.Username,
		Password: request.Password,
		Email:    request.Email,
	}

	var resp CreateAccountResponse

	err := r.tx.WithTransaction(ctx, func(ctx context.Context) error {
		acc, err := r.accountService.CreateAccount(ctx, accReq)
		if err != nil {
			return err
		}

		slog.Info("account created successfully", "name", acc.Username)

		ply, err := r.playerService.CreatePlayer(ctx, acc.Username, acc.Id)
		if err != nil {
			return err
		}

		slog.Info("player created successfully", "name", ply.Name)

		resp = CreateAccountResponse{
			AccountID:   acc.Id,
			AccountName: acc.Username,
			CreatedAt:   acc.CreatedAt,
			Email:       acc.Email.String(),
			User: SimpleUserResponse{
				UserID:   ply.Id,
				Username: ply.Name,
			},
		}

		return nil
	})

	if err != nil {
		return CreateAccountResponse{}, err
	}

	return resp, nil
}
