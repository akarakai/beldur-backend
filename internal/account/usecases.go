package account

import (
	"beldur/internal/auth"
	"beldur/internal/player"
	"beldur/pkg/db/postgres"
	"beldur/pkg/db/tx"
	"context"
	"errors"
	"log/slog"
)

// Registration is an USE CASE where an account is created along with a player of that account
type Registration struct {
	accountRepo     Repository
	uniquePlayerSvc *player.UniquePlayerService
	tx              tx.Transactor
	tokenIssuer     auth.TokenIssuer
}

// UsernamePasswordLogin is a login USE CASE
type UsernamePasswordLogin struct {
	accountRepo Repository
	tokenIssuer auth.TokenIssuer
}

func NewAccountRegistration(tx tx.Transactor, accountRepo Repository,
	uniquePlayerSvc *player.UniquePlayerService,
	tokenIssuer auth.TokenIssuer,
) *Registration {
	return &Registration{
		accountRepo:     accountRepo,
		uniquePlayerSvc: uniquePlayerSvc,
		tx:              tx,
		tokenIssuer:     tokenIssuer,
	}
}

func NewUsernamePasswordLogin(accountRepo Repository, tokenIssuer auth.TokenIssuer) *UsernamePasswordLogin {
	return &UsernamePasswordLogin{
		accountRepo: accountRepo,
		tokenIssuer: tokenIssuer,
	}
}

// RegisterAccount creates a new account and associates with it a player, all in a single transaction.
func (a *Registration) RegisterAccount(ctx context.Context, request CreateAccountRequest) (CreateAccountResponse, string, error) {
	newAcc, err := a.buildNewAccountFromRequest(request)
	if err != nil {
		return CreateAccountResponse{}, "", err
	}

	// Build player before tx (no DB reads)
	newPl, err := a.buildPlayer(newAcc.Username)
	if err != nil {
		return CreateAccountResponse{}, "", err
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

		// 2) Save player (service handles retries / unique violation)
		savedPl, err := a.uniquePlayerSvc.CreateUniquePlayer(ctx, newPl, newAcc.Id)
		if err != nil {
			return err
		}
		newPl = savedPl

		return nil
	})

	if err != nil {
		slog.Error("failed to register account", "username", request.Username, "error", err)
		return CreateAccountResponse{}, "", err
	}

	// generate token
	token, err := a.tokenIssuer.Issue(ctx, auth.Claims{
		Subject:   newAcc.Username,
		AccountID: newAcc.Id,
	})
	if err != nil {
		slog.Error("failed to issue token", "username", newAcc.Username, "error", err)
		return CreateAccountResponse{}, "", err
	}

	return CreateAccountResponse{
		AccountID:   newAcc.Id,
		AccountName: newAcc.Username,
		Email:       newAcc.Email.String(),
		CreatedAt:   newAcc.CreatedAt,
		Player: PlayerCreateResponse{
			PlayerID: newPl.Id,
			Name:     newPl.Name,
		},
	}, token, nil
}

func (a *Registration) buildNewAccountFromRequest(req CreateAccountRequest) (*Account, error) {
	hashedPass, err := HashPassword(req.Password)
	if err != nil {
		slog.Error("failed to hash password", "err", err)
		return nil, err
	}

	if req.Email == "" {
		newAcc, err := New(req.Username, hashedPass)
		if err != nil {
			slog.Error("failed to create new account", "err", err)
			return nil, err
		}
		return newAcc, nil
	}

	newAcc, err := New(req.Username, hashedPass, WithEmail(req.Email))
	if err != nil {
		slog.Error("failed to create new account", "err", err)
		return nil, err
	}
	return newAcc, nil
}

func (a *Registration) buildPlayer(accountName string) (*player.Player, error) {
	pl, err := player.New(accountName)
	if err != nil {
		slog.Error("failed to create new player", "err", err)
		return nil, err
	}
	return pl, nil
}

// Login returns nil and a new JWT authentication token if login is successful.
// Doesn't run in a transaction because readonly
// On login update the last access.
func (l *UsernamePasswordLogin) Login(ctx context.Context, request UsernamePasswordLoginRequest) (string, error) {
	username, pass := request.Username, request.Password

	acc, err := l.accountRepo.FindByUsername(ctx, username)
	if err != nil {
		slog.Info("failed to find account", "username", username, "error", err)
		return "", ErrDatabaseError // or wrap/map
	}

	if acc == nil || !CheckPasswordHash(acc.Password, pass) {
		return "", ErrInvalidCredentials
	}

	// login is successful, update the last access. This should never give an error
	if err := l.accountRepo.UpdateLastAccess(ctx, acc.Id); err != nil {
		slog.Error("failed to update last access", "error", err)
	}

	token, err := l.tokenIssuer.Issue(ctx, auth.Claims{
		Subject:   acc.Username,
		AccountID: acc.Id,
	})
	if err != nil {
		slog.Error("failed to issue token", "error", err)
		return "", err
	}
	return token, nil
}
