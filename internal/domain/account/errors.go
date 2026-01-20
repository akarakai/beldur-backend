package account

import "errors"

var (
	ErrInvalidUsername = errors.New("invalid account username")
	ErrInvalidPassword = errors.New("invalid account raw password")
	ErrHashing         = errors.New("could not hash the password")

	// service domain errors
	ErrUsernameAlreadyTaken = errors.New("username already taken")
)
