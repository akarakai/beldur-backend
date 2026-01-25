package campaign

import (
	"beldur/internal/id"
	"time"
)

const (
	MaxNameCharacters        = 50
	MaxDescriptionCharacters = 200

	MinPlayersNumber = 2   // master + a player
	MaxPlayersNumber = 100 // TODO check requirements
)

type Campaign struct {
	id          id.CampaignId
	name        string
	description string
	createdAt   time.Time
	// nil if not started
	startedAt *time.Time
	// nil if not finished
	finishedAt *time.Time
	status     StatusCampaign
	master     id.PlayerId
	// all the players of the campaign, included the master
	players map[id.PlayerId]struct{}
}

// New creates a new campaign. It only creates one, but doesn't start it.
// It is mandatory to know the master at the moment of creation.
func New(name string, description string, masterId id.PlayerId) (*Campaign, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}

	if err := validateDescription(description); err != nil {
		return nil, err
	}

	// TODO I do not know if master should be considered a player FOR NOW YES
	players := make(map[id.PlayerId]struct{})
	players[masterId] = struct{}{}

	return &Campaign{
		name:        name,
		description: description,
		createdAt:   time.Now(),
		startedAt:   nil,
		finishedAt:  nil,
		status:      StatusCreated,
		master:      masterId,
		players:     players,
	}, nil
}

// Start a new campaign. Only a created campaign can be started.
// An already started is idempotent.
func (c *Campaign) Start() error {
	// idempotent first
	if c.status == StatusStarted {
		return nil
	}

	if c.status == StatusFinished {
		return ErrCampaignFinished
	}

	if c.status == StatusCancelled {
		return ErrCampaignCancelled
	}

	if c.status != StatusCreated {
		return ErrCampaignNotCreated
	}

	if len(c.players) < MinPlayersNumber {
		return ErrNotEnoughPlayersToStart
	}

	now := time.Now()
	c.status = StatusStarted
	c.startedAt = &now

	return nil
}

// TODO
func (c *Campaign) CanBeJoined() bool {
	if c.status == StatusCreated {
		return true
	}
	if len(c.players) < MaxPlayersNumber {
		return true
	}
	return false
}

func (c *Campaign) Finish() error {
	if c.status == StatusFinished {
		return ErrCampaignFinished
	}

	if c.status == StatusCancelled {
		return ErrCampaignCancelled
	}

	if c.status != StatusStarted {
		return ErrCampaignNotStarted
	}

	now := time.Now()
	c.status = StatusFinished
	c.finishedAt = &now
	return nil
}

func (c *Campaign) Cancel() error {
	if c.status == StatusCancelled {
		return ErrCampaignCancelled
	}

	if c.status == StatusFinished {
		return ErrCampaignFinished
	}

	if c.status == StatusStarted {
		return ErrCampaignAlreadyStarted
	}

	if c.status != StatusCreated {
		return ErrCampaignNotCreated
	}

	now := time.Now()
	c.status = StatusCancelled
	c.finishedAt = &now
	return nil
}

func (c *Campaign) AddPlayer(playerId id.PlayerId) error {
	if c.status == StatusFinished {
		return ErrCampaignFinished
	}

	if c.status == StatusCancelled {
		return ErrCampaignCancelled
	}

	if _, exists := c.players[playerId]; exists {
		return ErrPlayerAlreadyInCampaign
	}

	c.players[playerId] = struct{}{}
	return nil
}

func validateName(name string) error {
	if len(name) > MaxNameCharacters {
		return ErrInvalidCampaignName
	}
	return nil
}

func validateDescription(description string) error {
	if len(description) > MaxDescriptionCharacters {
		return ErrInvalidCampaignDescription
	}
	return nil
}
