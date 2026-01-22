package campaign

import (
	"beldur/pkg/httperr"
	"errors"
	"net/http"
)

var (
	ErrInvalidCampaignName        = errors.New("invalid campaign name")
	ErrInvalidCampaignDescription = errors.New("invalid campaign description")
	ErrPlayerAlreadyInCampaign    = errors.New("player already in campaign")
	ErrCampaignFinished           = errors.New("campaign is finished")
	ErrCampaignCancelled          = errors.New("campaign is cancelled")
	ErrCampaignNotCreated         = errors.New("campaign is still not created")
	ErrCampaignNotStarted         = errors.New("campaign is still not started")
	ErrCampaignAlreadyStarted     = errors.New("campaign is already started")
	ErrNotEnoughPlayersToStart    = errors.New("not enough players to start the campaign")
)

func NewCampaignApiErrorManager() *httperr.Manager {
	mng := httperr.NewManager()

	mng.Add(ErrInvalidCampaignName, httperr.Mapped{
		Status:  http.StatusBadRequest,
		Code:    "invalid_campaign_name",
		Message: ErrInvalidCampaignName.Error(),
	})

	mng.Add(ErrInvalidCampaignDescription, httperr.Mapped{
		Status:  http.StatusBadRequest,
		Code:    "invalid_campaign_description",
		Message: ErrInvalidCampaignDescription.Error(),
	})

	mng.Add(ErrPlayerAlreadyInCampaign, httperr.Mapped{
		Status:  http.StatusConflict,
		Code:    "player_already_in_campaign",
		Message: ErrPlayerAlreadyInCampaign.Error(),
	})

	mng.Add(ErrCampaignFinished, httperr.Mapped{
		Status:  http.StatusConflict,
		Code:    "campaign_finished",
		Message: ErrCampaignFinished.Error(),
	})

	mng.Add(ErrCampaignCancelled, httperr.Mapped{
		Status:  http.StatusConflict,
		Code:    "campaign_cancelled",
		Message: ErrCampaignCancelled.Error(),
	})

	mng.Add(ErrCampaignNotCreated, httperr.Mapped{
		Status:  http.StatusConflict,
		Code:    "campaign_not_created",
		Message: ErrCampaignNotCreated.Error(),
	})

	mng.Add(ErrCampaignNotStarted, httperr.Mapped{
		Status:  http.StatusConflict,
		Code:    "campaign_not_started",
		Message: ErrCampaignNotStarted.Error(),
	})

	mng.Add(ErrCampaignAlreadyStarted, httperr.Mapped{
		Status:  http.StatusConflict,
		Code:    "campaign_already_started",
		Message: ErrCampaignAlreadyStarted.Error(),
	})

	mng.Add(ErrNotEnoughPlayersToStart, httperr.Mapped{
		Status:  http.StatusBadRequest,
		Code:    "not_enough_players_to_start",
		Message: ErrNotEnoughPlayersToStart.Error(),
	})

	return mng
}
