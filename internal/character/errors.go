package character

import (
	"beldur/pkg/httperr"
	"errors"
	"net/http"
)

// Item error
var (
	ErrNegativeAbility = errors.New("negative ability")

	// duplication with campaign?
	ErrCampaignNotFound               = errors.New("campaign not found")
	ErrCampaignHasAnotherMaster error = errors.New("campaign has another master")
)

func NewCharacterApiErrorManager() *httperr.Manager {
	mng := httperr.NewManager()

	mng.Add(ErrCampaignHasAnotherMaster, httperr.Mapped{
		Status:  http.StatusConflict,
		Code:    "campaign_has_another_master",
		Message: ErrCampaignHasAnotherMaster.Error(),
	})

	mng.Add(ErrCampaignNotFound, httperr.Mapped{
		Status:  http.StatusNotFound,
		Code:    "campaign_not_found",
		Message: ErrCampaignNotFound.Error(),
	})

	return mng
}
