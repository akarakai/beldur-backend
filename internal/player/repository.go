package player

import (
	"beldur/internal/id"
	"context"
)

type Saver interface {
	Save(ctx context.Context, player *Player, accountId id.AccountId) (*Player, error)
}

type Finder interface {
	FindByUsername(ctx context.Context, username string) (*Player, error)
	FindById(ctx context.Context, playerId id.PlayerId) (*Player, error)
	FindByAccountId(ctx context.Context, accountId id.AccountId) (*Player, error)
}

type Updater interface{}
