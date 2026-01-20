package account

import "context"

type Repository interface {
	Save(ctx context.Context, account *Account) (*Account, error)
	FindByUsername(ctx context.Context, username string) (*Account, error)
	FindById(ctx context.Context, accountId int) (*Account, error)
}
