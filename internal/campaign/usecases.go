package campaign

import (
	"beldur/internal/id"
	"beldur/pkg/db/tx"
	"context"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2/log"
)

type UseCase struct {
	campaignRepo Repository
	tx           tx.Transactor
}

func NewUseCase(campaignRepo Repository, tx tx.Transactor) *UseCase {
	return &UseCase{
		campaignRepo: campaignRepo,
		tx:           tx,
	}
}

// CreateNewCampaign creates a new fresh campaign without starting it.
// A player that creates a campaign become automatically the master of it.
// The existence of the master is not checked because it is taken from the authentication
func (uc *UseCase) CreateNewCampaign(ctx context.Context, req CreationRequest, masterId id.PlayerId) (CreationResponse, error) {
	var resp CreationResponse

	c, err := New(req.Name, req.Description, masterId)
	if err != nil {
		slog.Info("failed to create new campaign", "error", err)
		return CreationResponse{}, err // I think its safe to return this domain error to the user
	}

	code := generateAccessCode()

	err = uc.tx.WithTransaction(ctx, func(ctx context.Context) error {
		if err := uc.campaignRepo.Save(ctx, c, code); err != nil {
			slog.Error("failed to save new campaign", "error", err)
			return err
		}
		resp = CreationResponse{
			ID:          int(c.id),
			Name:        c.name,
			Description: c.description,
			Status:      string(c.status),
			CreatedAt:   c.createdAt,
			MasterID:    int(c.master),
			AccessCode:  code,
		}
		return nil
	})
	if err != nil {
		slog.Error("failed to save new campaign", "error", err)
		return CreationResponse{}, err
	}
	return resp, nil
}

// JoinCampaign lets a player join a campaign that has not started yet.
// The player must provide a valid authentication code generated at campaign creation.
func (uc *UseCase) JoinCampaign(ctx context.Context, req JoinRequest, campaignId id.CampaignId, playerId id.PlayerId) (JoinResponse, error) {
	var resp JoinResponse

	authCode := req.Code
	authCode = strings.TrimSpace(authCode)
	authCode = strings.ToUpper(authCode)

	err := uc.tx.WithTransaction(ctx, func(ctx context.Context) error {
		c, err := uc.campaignRepo.FindById(ctx, campaignId)
		if err != nil {
			slog.Info("no campaign found", "campaign_id", campaignId)
			return ErrCampaignNotFound
		}
		// check if authcode is the same
		codeDb, err := uc.campaignRepo.FindAuthCode(ctx, c.id)
		if err != nil {
			slog.Error("failed to find auth code", "campaign_id", c.id, "error", err)
			return err
		}
		if authCode != codeDb {
			return ErrWrongAccessCode
		}
		if err := c.AddPlayer(playerId); err != nil {
			return err
		}
		if err := uc.campaignRepo.Update(ctx, c); err != nil {
			slog.Error("failed to update campaign", "error", err)
			return err
		}
		resp = JoinResponse{
			ID:          int(c.id),
			Name:        c.name,
			Description: c.description,
			Status:      string(c.status),
			CreatedAt:   c.createdAt,
		}
		return nil
	})
	if err != nil {
		return JoinResponse{}, err
	}
	return resp, nil
}

// SearchCampaign gives back a list of campaigns, filtering is now not present
// Only the full list of the campaign in the database is given
func (uc *UseCase) SearchCampaign(ctx context.Context) ([]SimpleCampaignInfoResponse, error) {
	campaigns, err := uc.campaignRepo.FindAll(ctx)
	if err != nil {
		log.Error("failed to find campaigns from the database", "error", err)
		return []SimpleCampaignInfoResponse{}, err
	}

	cRespList := make([]SimpleCampaignInfoResponse, len(campaigns))
	for i, c := range campaigns {
		cResp := SimpleCampaignInfoResponse{
			ID:            int(c.id),
			Name:          c.name,
			Description:   c.description,
			Status:        string(c.status),
			CreatedAt:     c.createdAt,
			StartedAt:     c.startedAt,
			NumberPlayers: len(c.players) - 1,
			CanBeJoined:   c.CanBeJoined(),
		}
		cRespList[i] = cResp
	}
	return cRespList, nil
}
