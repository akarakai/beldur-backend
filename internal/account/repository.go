package account

import (
	"beldur/internal/id"
	"context"
)

type Repository interface {
	Save(ctx context.Context, account *Account) (*Account, error)
	FindByUsername(ctx context.Context, username string) (*Account, error)
	FindById(ctx context.Context, accountId id.AccountId) (*Account, error)
	UpdateLastAccess(ctx context.Context, accountId id.AccountId) error
}
