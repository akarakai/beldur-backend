package middleware

import (
	"beldur/internal/id"
	"context"

	"github.com/gofiber/fiber/v2"
)

type CampaignMasterChecker interface {
	IsMasterOfCampaign(ctx context.Context, campaignId id.CampaignId, playerId id.PlayerId) (bool, error)
}

func CampaignMiddleware(service CampaignMasterChecker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return nil
	}
}
