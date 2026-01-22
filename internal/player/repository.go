package player

import (
	"beldur/internal/id"
	"context"
)

type Repository interface {
	Save(ctx context.Context, player *Player, accountId id.AccountId) (*Player, error)
	FindByUsername(ctx context.Context, username string) (*Player, error)
	FindById(ctx context.Context, playerId id.PlayerId) (*Player, error)
	FindByAccountId(ctx context.Context, accountId id.AccountId) (*Player, error)
}
