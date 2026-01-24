package campaign

import (
	"beldur/internal/id"
	"context"
)

type Finder interface {
	FindById(ctx context.Context, campaignId id.CampaignId) (*Campaign, error)
	FindAuthCode(ctx context.Context, campaignId id.CampaignId) (string, error)
	FindAll(ctx context.Context) ([]*Campaign, error)
}

type Updater interface {
	Update(ctx context.Context, campaign *Campaign) error
}

type Saver interface {
	Save(ctx context.Context, campaign *Campaign, accessCode string) error
}
