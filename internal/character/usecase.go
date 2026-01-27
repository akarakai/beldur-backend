package character

import (
	"beldur/internal/id"
	"context"
	"errors"
)

type CreateUseCase struct {
	campaignFinder CampaignFinder
	characterSaver Saver
}

func NewCreateUseCase(campaignFinder CampaignFinder, characterSaver Saver) *CreateUseCase {
	return &CreateUseCase{
		campaignFinder: campaignFinder,
		characterSaver: characterSaver,
	}
}

// CreateNPC creates a basic NPC character given some data as the abilities.
// the created NPC has no equipment and should be added with other requests
// via the UC create item, these items should then be added via a request.
// Only a master of the campaign can create the NPC.
func (uc *CreateUseCase) CreateNPC(
	ctx context.Context,
	req CreateCharacterRequest,
	campaignId id.CampaignId,
	masterId id.PlayerId) (CreateCharacterResponse, error) {
	abilities := uc.getAbilities(req)

	ch := New(req.Name, req.Description, WithAbilities(abilities))

	// get he campaign
	camp, err := uc.campaignFinder.FindById(ctx, campaignId)
	if err != nil {
		return CreateCharacterResponse{}, errors.Join(ErrCampaignNotFound, err)
	}

	// Maybe as a middleware, soft authentication. Because the role is based on campaign, not on account
	if !camp.IsMaster(masterId) {
		return CreateCharacterResponse{}, ErrCampaignHasAnotherMaster
	}

	// Nothing to do here... NPC is created, now I have to save him in the repository
	// maybe I will need for a Transactor if multiple queries are needed
	if err := uc.characterSaver.SaveNPC(ctx, ch, camp.Id(), masterId); err != nil {
		return CreateCharacterResponse{}, errors.New("failed to save character")
	}

	return CreateCharacterResponse{
		Id:          int(ch.id),
		Name:        ch.name,
		Description: ch.description,
		CampaignId:  int(camp.Id()),
		Abilities:   req.Abilities, // taken from the request
	}, nil
}

// CreatePlayerCharacter creates a character from the campaign. Each player creates a character for himself.
// One character for player for campaign
func (uc *CreateUseCase) CreatePlayerCharacter(req CreateCharacterRequest, campaignId id.CharacterId, playerId id.PlayerId) string {

	return ""
}

func (uc *CreateUseCase) getAbilities(req CreateCharacterRequest) Abilities {
	return NewAbilities(req.Abilities.Strength, req.Abilities.Dexterity, req.Abilities.Constitution, req.Abilities.Intelligence, req.Abilities.Wisdom, req.Abilities.Charisma)
}
