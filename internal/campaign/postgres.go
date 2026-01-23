package campaign

import (
	"beldur/internal/id"
	"beldur/pkg/db/postgres"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type PostgresRepository struct {
	q postgres.QuerierProvider
}

func NewPostgresRepository(q postgres.QuerierProvider) *PostgresRepository {
	return &PostgresRepository{q: q}
}

func (p *PostgresRepository) Save(ctx context.Context, c *Campaign, code string) error {
	const sqlCampaign = `
		INSERT INTO campaigns (name, description, created_at, status, master_id, access_code)
		VALUES ($1, $2, $3, $4, $5, $6)
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
		code,
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

func (p *PostgresRepository) FindById(ctx context.Context, campaignId id.CampaignId) (*Campaign, error) {
	const sql = `
		SELECT 
		    c.campaign_id,
		    c.name,
		    c.description,
		    c.created_at,
		    c.started_at,
		    c.finished_at,
		    c.status,
		    c.master_id,
		    cp.player_id,
		    cp.is_master
		FROM campaigns c
		INNER JOIN campaigns_players cp ON c.campaign_id = cp.campaign_id
		WHERE c.campaign_id = $1
	`

	rows, err := p.q(ctx).Query(ctx, sql, int(campaignId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaign *Campaign

	for rows.Next() {
		var (
			campaignID  int
			name        string
			description string
			createdAt   time.Time
			startedAt   *time.Time
			finishedAt  *time.Time
			status      StatusCampaign
			masterID    int
			playerID    int
			isMaster    bool
		)

		if err := rows.Scan(
			&campaignID,
			&name,
			&description,
			&createdAt,
			&startedAt,
			&finishedAt,
			&status,
			&masterID,
			&playerID,
			&isMaster,
		); err != nil {
			return nil, err
		}

		// initialize campaign
		if campaign == nil {
			campaign = &Campaign{
				id:          id.CampaignId(campaignID),
				name:        name,
				description: description,
				createdAt:   createdAt,
				startedAt:   startedAt,
				finishedAt:  finishedAt,
				status:      status,
				master:      id.PlayerId(masterID),
				players:     make(map[id.PlayerId]struct{}),
			}
		}
		campaign.players[id.PlayerId(playerID)] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if campaign == nil {
		return nil, postgres.ErrNoRowFound
	}
	return campaign, nil
}

func (p *PostgresRepository) FindAuthCode(ctx context.Context, campaignId id.CampaignId) (string, error) {
	const sql = `
		SELECT access_code
		FROM campaigns
		WHERE campaign_id = $1
`
	var code string

	if err := p.q(ctx).QueryRow(ctx, sql, campaignId).Scan(&code); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", postgres.ErrNoRowFound
		}
	}
	return code, nil
}

func (p *PostgresRepository) Update(ctx context.Context, campaign *Campaign) error {
	const sqlUpdateCampaign = `
        UPDATE campaigns 
        SET name = $1,
            description = $2,
            started_at = $3,
            finished_at = $4,
            status = $5
        WHERE campaign_id = $6
    `

	if _, err := p.q(ctx).Exec(ctx,
		sqlUpdateCampaign,
		campaign.name,
		campaign.description,
		campaign.startedAt,
		campaign.finishedAt,
		string(campaign.status),
		campaign.id,
	); err != nil {
		return err
	}

	const sqlInsertPlayer = `
        INSERT INTO campaigns_players (campaign_id, player_id, is_master)
        VALUES ($1, $2, $3)
        ON CONFLICT (campaign_id, player_id) DO NOTHING
    `

	for playerID := range campaign.players {
		if _, err := p.q(ctx).Exec(ctx,
			sqlInsertPlayer,
			campaign.id,
			playerID,
			playerID == campaign.master,
		); err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresRepository) FindAll(ctx context.Context) ([]*Campaign, error) {
	const sql = `
		SELECT 
			c.campaign_id,
			c.name,
			c.description,
			c.created_at,
			c.started_at,
			c.finished_at,
			c.status,
			c.master_id,
			cp.player_id
		FROM campaigns c
		LEFT JOIN campaigns_players cp ON c.campaign_id = cp.campaign_id
		ORDER BY c.campaign_id
	`

	rows, err := p.q(ctx).Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byID := make(map[id.CampaignId]*Campaign)
	order := make([]id.CampaignId, 0)

	for rows.Next() {
		var (
			campaignID  int
			name        string
			description string
			createdAt   time.Time
			startedAt   *time.Time
			finishedAt  *time.Time
			status      StatusCampaign
			masterID    int
			playerID    *int
		)

		if err := rows.Scan(
			&campaignID,
			&name,
			&description,
			&createdAt,
			&startedAt,
			&finishedAt,
			&status,
			&masterID,
			&playerID,
		); err != nil {
			return nil, err
		}

		cid := id.CampaignId(campaignID)

		c, ok := byID[cid]
		if !ok {
			c = &Campaign{
				id:          cid,
				name:        name,
				description: description,
				createdAt:   createdAt,
				startedAt:   startedAt,
				finishedAt:  finishedAt,
				status:      status,
				master:      id.PlayerId(masterID),
				players:     make(map[id.PlayerId]struct{}),
			}
			byID[cid] = c
			order = append(order, cid)
		}

		// add player if present (LEFT JOIN can be NULL)
		if playerID != nil {
			c.players[id.PlayerId(*playerID)] = struct{}{}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// build result slice in order
	campaigns := make([]*Campaign, 0, len(order))
	for _, cid := range order {
		campaigns = append(campaigns, byID[cid])
	}

	if len(campaigns) == 0 {
		return []*Campaign{}, nil
	}

	return campaigns, nil
}
