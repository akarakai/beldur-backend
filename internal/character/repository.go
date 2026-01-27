package character

import (
	"beldur/internal/campaign"
	"beldur/internal/id"
	"context"
)

type CampaignFinder interface {
	FindById(ctx context.Context, campaignId id.CampaignId) (*campaign.Campaign, error)
}

type Saver interface {
	SavePlayerCharacter(ctx context.Context, character *Character, campaignId id.CampaignId, playerId id.PlayerId) error
	SaveNPC(ctx context.Context, character *Character, campaignId id.CampaignId, masterId id.PlayerId) error
}
