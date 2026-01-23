package campaign

import (
	"beldur/internal/id"
	"context"
)

type Repository interface {
	// Save mutates campaign by adding id and other fields after persistence
	Save(ctx context.Context, campaign *Campaign, accessCode string) error
	FindById(ctx context.Context, campaignId id.CampaignId) (*Campaign, error)
	FindAuthCode(ctx context.Context, campaignId id.CampaignId) (string, error)
	Update(ctx context.Context, campaign *Campaign) error
	FindAll(ctx context.Context) ([]*Campaign, error)
}
