package player

import "context"

type Repository interface {
	Save(ctx context.Context, player *Player, accountId int) (*Player, error)
	FindByUsername(ctx context.Context, username string) (*Player, error)
	FindById(ctx context.Context, playerId int) (*Player, error)
}
