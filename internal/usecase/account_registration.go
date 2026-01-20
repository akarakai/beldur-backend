package usecase

import (
	"beldur/internal/db/postgres"
	"beldur/internal/db/tx"
	"beldur/internal/domain/account"
	"beldur/internal/domain/account/hash"
	"beldur/internal/domain/player"
	"context"
	"errors"
	"fmt"
	"log/slog"
)

type AccountRegistration struct {
	accountRepo account.Repository
	playerRepo  player.Repository
	tx          tx.Transactor
}

func NewAccountRegistration(tx tx.Transactor, accountRepo account.Repository, playerRepo player.Repository) *AccountRegistration {
	return &AccountRegistration{
		accountRepo: accountRepo,
		playerRepo:  playerRepo,
		tx:          tx,
	}
}

// RegisterAccount creates a new account and associates with it a player, all in a single transaction.
func (a *AccountRegistration) RegisterAccount(ctx context.Context, request CreateAccountRequest) (CreateAccountResponse, error) {
	newAcc, err := a.buildNewAccountFromRequest(request)
	if err != nil {
		return CreateAccountResponse{}, err
	}

	// Build player before tx (no DB reads)
	newPl, err := a.buildPlayer(newAcc.Username)
	if err != nil {
		return CreateAccountResponse{}, err
	}

	err = a.tx.WithTransaction(ctx, func(ctx context.Context) error {
		// 1) Save account (source of truth is DB unique constraint)
		savedAcc, err := a.accountRepo.Save(ctx, newAcc)
		if err != nil {
			if errors.Is(err, postgres.ErrUniqueValueViolation) {
				slog.Info("account unique constraint violation", "username", newAcc.Username)
				return ErrAccountNameAlreadyTaken
			}

			slog.Error("failed to save new account", "username", newAcc.Username, "error", err)
			return errors.Join(ErrDatabaseError, err)
		}
		newAcc = savedAcc

		savedPl, err := a.persistPlayerUntilValid(ctx, newPl, newAcc.Id)
		if err != nil {
			return err
		}
		newPl = savedPl

		return nil
	})

	if err != nil {
		slog.Error("failed to register account", "username", request.Username, "error", err)
		return CreateAccountResponse{}, err
	}

	return CreateAccountResponse{
		AccountID:   newAcc.Id,
		AccountName: newAcc.Username,
		Email:       newAcc.Email.String(),
		CreatedAt:   newAcc.CreatedAt,
		Player: SimplePlayerResponse{
			UserID: newPl.Id,
			Name:   newPl.Name,
		},
	}, nil
}

// persistPlayerUntilValid attempts to persist a player
// If the name is already taken (unique violation), it generates a new name and retries.
func (a *AccountRegistration) persistPlayerUntilValid(ctx context.Context, newPl *player.Player, accountID int) (*player.Player, error) {
	const maxTrials = 3

	originalName := newPl.Name

	for attempt := 0; attempt < maxTrials; attempt++ {
		savedPl, err := a.playerRepo.Save(ctx, newPl, accountID)
		if err == nil {
			return savedPl, nil
		}

		if !errors.Is(err, postgres.ErrUniqueValueViolation) {
			slog.Error("failed to save new player", "error", err)
			return nil, errors.Join(ErrDatabaseError, err)
		}

		newName := fmt.Sprintf("%s_%d", originalName, attempt+1)
		slog.Info("player name already taken, retrying",
			"playerName", newPl.Name,
			"newPlayerName", newName,
			"attempt", attempt+1,
			"maxTrials", maxTrials,
		)

		if err := newPl.ChangeName(newName); err != nil {
			slog.Error("failed to change player name", "error", err)
			return nil, err
		}
	}

	slog.Error("failed to save new player after retries",
		"originalName", originalName,
		"finalName", newPl.Name,
		"maxTrials", maxTrials,
		"error", ErrPlayerNameAlreadyTaken,
	)
	return nil, ErrPlayerNameAlreadyTaken
}

func (a *AccountRegistration) buildNewAccountFromRequest(req CreateAccountRequest) (*account.Account, error) {
	hashedPass, err := hash.HashPassword(req.Password)
	if err != nil {
		slog.Error("failed to hash password", "err", err)
		return nil, err
	}

	if req.Email == "" {
		newAcc, err := account.New(req.Username, hashedPass)
		if err != nil {
			slog.Error("failed to create new account", "err", err)
			return nil, err
		}
		return newAcc, nil
	}

	newAcc, err := account.New(req.Username, hashedPass, account.WithEmail(req.Email))
	if err != nil {
		slog.Error("failed to create new account", "err", err)
		return nil, err
	}
	return newAcc, nil
}

func (a *AccountRegistration) buildPlayer(accountName string) (*player.Player, error) {
	pl, err := player.New(accountName)
	if err != nil {
		slog.Error("failed to create new player", "err", err)
		return nil, err
	}
	return pl, nil
}
