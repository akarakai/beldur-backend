package player

import "errors"

var (
	ErrInvalidPlayerName      = errors.New("invalid player name")
	ErrCouldNotSavePlayer     = errors.New("could not save player with given username")
	ErrPlayerNameAlreadyTaken = errors.New("player name already taken")
)
