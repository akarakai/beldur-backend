package account

import (
	"beldur/internal/id"
	"context"
)

type Finder interface {
	FindByUsername(ctx context.Context, username string) (*Account, error)
	FindById(ctx context.Context, accountId id.AccountId) (*Account, error)
}

type Updater interface {
	UpdateLastAccess(ctx context.Context, accountId id.AccountId) error
	Update(ctx context.Context, account *Account) error
}

type Saver interface {
	Save(ctx context.Context, account *Account) (*Account, error)
}
