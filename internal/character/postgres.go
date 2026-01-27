package character

import (
	"beldur/internal/id"
	"beldur/pkg/db/postgres"
	"context"
)

type PostgresRepository struct {
	q postgres.QuerierProvider
}

func NewPostgresRepository(q postgres.QuerierProvider) *PostgresRepository {
	return &PostgresRepository{q: q}
}

func (p *PostgresRepository) SavePlayerCharacter(
	ctx context.Context,
	c *Character,
	campaignId id.CampaignId,
	playerId id.PlayerId,
) error {
	return p.save(ctx, c, campaignId, playerId, false)
}

func (p *PostgresRepository) SaveNPC(
	ctx context.Context,
	c *Character,
	campaignId id.CampaignId,
	masterId id.PlayerId,
) error {
	return p.save(ctx, c, campaignId, masterId, true)
}

func (p *PostgresRepository) save(
	ctx context.Context,
	c *Character,
	campaignId id.CampaignId,
	masterId id.PlayerId,
	isNPC bool,
) error {
	const query = `
		INSERT INTO characters 
		    (campaign_id, player_id, name, description, 
		     base_strength, base_dexterity, base_constitution, 
		     base_intelligence, base_wisdom, base_charisma, is_npc)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING character_id
	`

	row := p.q(ctx).QueryRow(ctx, query,
		int(campaignId),
		int(masterId),
		c.name,
		c.description,
		c.abilities.Get(AbilityStrength),
		c.abilities.Get(AbilityDexterity),
		c.abilities.Get(AbilityConstitution),
		c.abilities.Get(AbilityIntelligence),
		c.abilities.Get(AbilityWisdom),
		c.abilities.Get(AbilityCharisma),
		isNPC,
	)
	var characterID int
	if err := row.Scan(&characterID); err != nil {
		return err
	}
	c.id = id.CharacterId(characterID)
	return nil
}
