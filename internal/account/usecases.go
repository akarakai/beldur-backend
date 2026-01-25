package account

import (
	"beldur/internal/id"
	"beldur/internal/player"
	"beldur/pkg/auth"
	"beldur/pkg/db/postgres"
	"beldur/pkg/db/tx"
	"beldur/pkg/logger"
	"context"
	"errors"
)

type UniquePlayerCreator interface {
	CreateUniquePlayer(ctx context.Context, pl *player.Player, accId id.AccountId) (*player.Player, error)
}

// Registration is an USE CASE where an account is created along with a player of that account
type Registration struct {
	accSaver        Saver
	uniquePlayerSvc UniquePlayerCreator
	tx              tx.Transactor
	tokenIssuer     auth.TokenIssuer
}

// UsernamePasswordLogin is a login USE CASE
type UsernamePasswordLogin struct {
	accFinder    Finder
	accUpdater   Updater
	playerFinder player.Finder
	tokenIssuer  auth.TokenIssuer
}

type Management struct {
	accFinder  Finder
	accUpdater Updater
}

func NewAccountManagement(accFinder Finder, accUpdater Updater) *Management {
	return &Management{
		accFinder:  accFinder,
		accUpdater: accUpdater,
	}
}

// UpdateAccount updates the account of accountId
func (uc *Management) UpdateAccount(ctx context.Context, req UpdateAccountRequest, accountId id.AccountId) (UpdateAccountResponse, error) {
	acc, err := uc.accFinder.FindById(ctx, accountId)
	if err != nil {
		logger.Debug("could not find account by id", "err", err)
		return UpdateAccountResponse{}, err
	}
	em, err := NewEmail(req.Email)
	if err != nil {
		return UpdateAccountResponse{}, err
	}

	acc.UpdateEmail(em)

	if err := uc.accUpdater.Update(ctx, acc); err != nil {
		return UpdateAccountResponse{}, err
	}

	return UpdateAccountResponse{
		Email: acc.Email.String(),
	}, nil

}

func NewAccountRegistration(tx tx.Transactor, accSaver Saver,
	uniquePlayerSvc UniquePlayerCreator,
	tokenIssuer auth.TokenIssuer,
) *Registration {
	return &Registration{
		accSaver:        accSaver,
		uniquePlayerSvc: uniquePlayerSvc,
		tx:              tx,
		tokenIssuer:     tokenIssuer,
	}
}

func NewUsernamePasswordLogin(accFinder Finder, accUpdater Updater, playerFinder player.Finder, tokenIssuer auth.TokenIssuer) *UsernamePasswordLogin {
	return &UsernamePasswordLogin{
		accFinder:    accFinder,
		accUpdater:   accUpdater,
		playerFinder: playerFinder,
		tokenIssuer:  tokenIssuer,
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
		savedAcc, err := a.accSaver.Save(ctx, newAcc)
		if err != nil {
			if errors.Is(err, postgres.ErrUniqueValueViolation) {
				logger.Debug("account unique constraint violation", "username", newAcc.Username)
				return ErrAccountNameAlreadyTaken
			}
			logger.Debug("failed to save new account", "err", err)
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
		logger.Debug("failed to register account", "err", err)
		return CreateAccountResponse{}, "", err
	}

	// generate token
	token, err := a.tokenIssuer.Issue(ctx, auth.Claims{
		Subject:  newAcc.Id,
		PlayerID: newPl.Id,
	})
	if err != nil {
		logger.Debug("failed to issue token", "err", err)
		return CreateAccountResponse{}, "", err
	}

	var emailVal *string
	if newAcc.Email != nil {
		emailStr := newAcc.Email.String()
		emailVal = &emailStr
	}

	return CreateAccountResponse{
		AccountID:   int(newAcc.Id),
		AccountName: newAcc.Username,
		Email:       emailVal,
		CreatedAt:   newAcc.CreatedAt,
		Player: PlayerCreateResponse{
			PlayerID: int(newPl.Id),
			Name:     newPl.Name,
		},
	}, token, nil
}

func (a *Registration) buildNewAccountFromRequest(req CreateAccountRequest) (*Account, error) {
	hashedPass, err := HashPassword(req.Password)
	if err != nil {
		logger.Debug("failed to hash password", "err", err)
		return nil, err
	}

	if req.Email == nil {
		newAcc, err := New(req.Username, hashedPass)
		if err != nil {
			logger.Debug("failed to create new account", "err", err)
			return nil, err
		}
		return newAcc, nil
	}

	em, err := NewEmail(*req.Email)
	if err != nil {
		return nil, err
	}

	newAcc, err := New(req.Username, hashedPass, WithEmail(em))
	if err != nil {
		logger.Debug("failed to create new account", "err", err)
		return nil, err
	}
	return newAcc, nil
}

func (a *Registration) buildPlayer(accountName string) (*player.Player, error) {
	pl, err := player.New(accountName)
	if err != nil {
		logger.Debug("failed to create new player", "err", err)
		return nil, err
	}
	return pl, nil
}

// Login returns nil and a new JWT authentication token if login is successful.
// Doesn't run in a transaction because readonly
// On login update the last access.
func (l *UsernamePasswordLogin) Login(ctx context.Context, request UsernamePasswordLoginRequest) (string, error) {
	username, pass := request.Username, request.Password

	acc, err := l.accFinder.FindByUsername(ctx, username)
	if err != nil {
		logger.Debug("failed to find account", "username", username, "error", err)
		return "", ErrDatabaseError // or wrap/map
	}

	if acc == nil || !CheckPasswordHash(acc.Password, pass) {
		return "", ErrInvalidCredentials
	}

	p, err := l.playerFinder.FindByAccountId(ctx, acc.Id)
	if err != nil {
		logger.Debug("failed to find player", "account", acc.Id, "error", err)
		return "", errors.Join(ErrDatabaseError, errors.New("failed to fetch the player even if account is found"))
	}

	// login is successful, update the last access. This should never give an error
	if err := l.accUpdater.UpdateLastAccess(ctx, acc.Id); err != nil {
		logger.Debug("failed to update last access", "error", err)
		return "", ErrDatabaseError
	}

	token, err := l.tokenIssuer.Issue(ctx, auth.Claims{
		Subject:  acc.Id,
		PlayerID: p.Id,
	})
	if err != nil {
		logger.Error("failed to issue token", err)
		return "", err
	}
	return token, nil
}
