package player

import (
	"beldur/internal/id"
	"beldur/pkg/db/postgres"
	"context"
	"errors"
	"fmt"
	"log/slog"
)

// UniquePlayerService is a domain service that makes sure to persist an user in its repository
// without violating the constriction of unique username
type UniquePlayerService struct {
	playerRepo Repository
}

func NewUniquePlayerService(playerRepo Repository) *UniquePlayerService {
	return &UniquePlayerService{playerRepo: playerRepo}
}

// CreateUniquePlayer attempts to persist a player.
// If the name is already taken (unique violation), it generates a new name and retries.
func (s *UniquePlayerService) CreateUniquePlayer(ctx context.Context, pl *Player, accountId id.AccountId) (*Player, error) {
	const maxTrials = 3

	originalName := pl.Name

	for attempt := 0; attempt < maxTrials; attempt++ {
		savedPl, err := s.playerRepo.Save(ctx, pl, accountId)
		if err == nil {
			return savedPl, nil
		}

		if !errors.Is(err, postgres.ErrUniqueValueViolation) {
			slog.Error("failed to save new player", "error", err)
			return nil, err
		}

		newName := fmt.Sprintf("%s_%d", originalName, attempt+1)
		slog.Info("player name already taken, retrying",
			"playerName", pl.Name,
			"newPlayerName", newName,
			"attempt", attempt+1,
			"maxTrials", maxTrials,
		)

		if err := pl.ChangeName(newName); err != nil {
			slog.Error("failed to change player name", "error", err)
			return nil, err
		}
	}

	slog.Error("failed to save new player after retries",
		"originalName", originalName,
		"finalName", pl.Name,
		"maxTrials", maxTrials,
		"error", ErrPlayerNameAlreadyTaken,
	)
	return nil, ErrPlayerNameAlreadyTaken
}
