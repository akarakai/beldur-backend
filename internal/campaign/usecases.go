package campaign

import (
	"beldur/internal/id"
	"beldur/pkg/db/tx"
	"context"
	"log/slog"
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

	err = uc.tx.WithTransaction(ctx, func(ctx context.Context) error {
		if err := uc.campaignRepo.Save(ctx, c); err != nil {
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
		}
		return nil
	})
	if err != nil {
		slog.Error("failed to save new campaign", "error", err)
		return CreationResponse{}, err
	}
	return resp, nil
}
