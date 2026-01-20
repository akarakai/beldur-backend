package postgres

import (
	"context"
	"errors"

	"beldur/internal/domain/player"

	"github.com/jackc/pgx/v5"
)

type PlayerRepository struct {
	q QuerierProvider
}

func NewPlayerRepository(q QuerierProvider) *PlayerRepository {
	return &PlayerRepository{q: q}
}

func (p *PlayerRepository) Save(ctx context.Context, pl *player.Player, accountId int) (*player.Player, error) {
	query := `
		INSERT INTO players (player_name, account_id)
		VALUES ($1, $2)
		RETURNING player_id, player_name
	`

	row := p.q(ctx).QueryRow(ctx, query, pl.Name, accountId)

	saved, err := p.scanPlayer(row)
	if err != nil {
		return nil, err
	}
	// INSERT ... RETURNING should always return a row
	if saved == nil {
		return nil, errors.New("insert player returned no row")
	}
	return saved, nil
}

func (p *PlayerRepository) FindByUsername(ctx context.Context, username string) (*player.Player, error) {
	query := `
		SELECT player_id, player_name
		FROM players
		WHERE player_name = $1
		LIMIT 1
	`

	row := p.q(ctx).QueryRow(ctx, query, username)
	return p.scanPlayer(row)
}

func (p *PlayerRepository) FindById(ctx context.Context, playerId int) (*player.Player, error) {
	query := `
		SELECT player_id, player_name
		FROM players
		WHERE player_id = $1
		LIMIT 1
	`

	row := p.q(ctx).QueryRow(ctx, query, playerId)
	return p.scanPlayer(row)
}

// scanPlayer translates DB row -> domain model.
// Returns (nil, nil) when no row is found.
func (p *PlayerRepository) scanPlayer(row pgx.Row) (*player.Player, error) {
	var (
		id   int
		name string
	)

	err := row.Scan(&id, &name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	pl, err := player.New(name)
	if err != nil {
		return nil, err
	}
	pl.Id = id
	return pl, nil
}
