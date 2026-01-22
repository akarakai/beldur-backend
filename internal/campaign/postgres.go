package campaign

import (
	"beldur/pkg/db/postgres"
	"context"
)

type PostgresRepository struct {
	q postgres.QuerierProvider
}

func NewPostgresRepository(q postgres.QuerierProvider) *PostgresRepository {
	return &PostgresRepository{q: q}
}

func (p *PostgresRepository) Save(ctx context.Context, c *Campaign) error {
	const sqlCampaign = `
		INSERT INTO campaigns (name, description, created_at, status, master_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING campaign_id;
	`

	const sqlCampaignPlayer = `
		INSERT INTO campaigns_players (campaign_id, player_id, is_master)
		VALUES ($1, $2, $3);
	`

	if err := p.q(ctx).QueryRow(ctx,
		sqlCampaign,
		c.name,
		c.description,
		c.createdAt,
		string(c.status),
		c.master,
	).Scan(&c.id); err != nil {
		return err
	}

	for playerID := range c.players {
		if _, err := p.q(ctx).Exec(ctx,
			sqlCampaignPlayer,
			c.id,
			playerID,
			playerID == c.master,
		); err != nil {
			return err
		}
	}
	return nil
}
