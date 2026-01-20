package usecase

import "errors"

var (
	ErrDatabaseError           = errors.New("database error when executing use case")
	ErrAccountNameAlreadyTaken = errors.New("account name already taken")
	ErrPlayerNameAlreadyTaken  = errors.New("player name already taken")
)
