package player

import (
	postgres2 "beldur/pkg/db/postgres"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresPlayerRepository struct {
	q postgres2.QuerierProvider
}

func NewPlayerRepository(q postgres2.QuerierProvider) *PostgresPlayerRepository {
	return &PostgresPlayerRepository{q: q}
}

func (p *PostgresPlayerRepository) Save(ctx context.Context, pl *Player, accountId int) (*Player, error) {
	query := `
		INSERT INTO players (player_name, account_id)
		VALUES ($1, $2)
		RETURNING player_id, player_name
	`

	row := p.q(ctx).QueryRow(ctx, query, pl.Name, accountId)

	saved, err := p.scanPlayer(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, postgres2.ErrUniqueValueViolation
		}
		return nil, err
	}
	// INSERT ... RETURNING should always return a row
	if saved == nil {
		return nil, errors.New("insert player returned no row")
	}
	return saved, nil
}

func (p *PostgresPlayerRepository) FindByUsername(ctx context.Context, username string) (*Player, error) {
	query := `
		SELECT player_id, player_name
		FROM players
		WHERE player_name = $1
		LIMIT 1
	`

	row := p.q(ctx).QueryRow(ctx, query, username)
	return p.scanPlayer(row)
}

func (p *PostgresPlayerRepository) FindById(ctx context.Context, playerId int) (*Player, error) {
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
func (p *PostgresPlayerRepository) scanPlayer(row pgx.Row) (*Player, error) {
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

	pl, err := New(name)
	if err != nil {
		return nil, err
	}
	pl.Id = id
	return pl, nil
}
