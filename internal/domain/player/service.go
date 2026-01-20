package player

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

const maxUsernameAttempts = 5

type Service struct {
	playerRepo Repository
}

func NewService(playerRepo Repository) *Service {
	if playerRepo == nil {
		panic("nil Repository")
	}

	return &Service{
		playerRepo: playerRepo,
	}
}

// CreatePlayer creates a player linked to an account.
// It starts with the account name as username and generates a new one if it already exists.
func (s *Service) CreatePlayer(ctx context.Context, playerName string, accountID int) (*Player, error) {
	username, err := s.generateUniqueUsername(ctx, playerName)
	if err != nil {
		slog.Error(
			"failed to generate unique username",
			"err", err,
			"playerName", playerName,
			"accountID", accountID,
		)
		return nil, err
	}

	player, err := New(username)
	if err != nil {
		slog.Error(
			"failed to construct new player",
			"err", err,
			"username", username,
		)
		return nil, err
	}

	savedPlayer, err := s.playerRepo.Save(ctx, player, accountID)
	if err != nil {
		slog.Error(
			"failed to save player in database",
			"err", err,
			"username", username,
			"accountID", accountID,
		)
		return nil, err
	}

	return savedPlayer, nil
}

func (s *Service) generateUniqueUsername(ctx context.Context, baseName string) (string, error) {
	for i := 0; i <= maxUsernameAttempts; i++ {
		username := baseName
		if i > 0 {
			username = fmt.Sprintf("%s%d", baseName, i)
		}

		player, err := s.playerRepo.FindByUsername(ctx, username)
		if err != nil {
			return "", err
		}

		if player == nil {
			return username, nil
		}
	}

	return "", errors.New("could not generate a unique username after maximum attempts")
}
