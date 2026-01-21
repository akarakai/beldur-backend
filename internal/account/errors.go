package account

import (
	"beldur/pkg/httperr"
	"errors"
	"net/http"
)

var (
	ErrInvalidUsername = errors.New("invalid account username")
	ErrInvalidPassword = errors.New("invalid account password")
	ErrHashing         = errors.New("could not hash the password")

	// service domain errors
)

var (
	ErrDatabaseError           = errors.New("database error when executing use case")
	ErrAccountNameAlreadyTaken = errors.New("account name already taken")
	ErrAccountDoesNotExist     = errors.New("account does not exist")
	ErrInvalidCredentials      = errors.New("invalid login credentials")
)

func NewAccountApiErrorManager() *httperr.Manager {
	mng := httperr.NewManager()

	mng.Add(ErrInvalidUsername, httperr.Mapped{
		Status:  http.StatusBadRequest,
		Code:    "invalid_username",
		Message: "Username is invalid",
	})

	mng.Add(ErrInvalidPassword, httperr.Mapped{
		Status:  http.StatusBadRequest,
		Code:    "invalid_password",
		Message: "Password is invalid",
	})

	mng.Add(ErrAccountNameAlreadyTaken, httperr.Mapped{
		Status:  http.StatusConflict,
		Code:    "username_taken",
		Message: "Username is already taken",
	})

	mng.Add(ErrInvalidCredentials, httperr.Mapped{
		Status:  http.StatusUnauthorized,
		Code:    "invalid_credentials",
		Message: "Invalid username or password",
	})

	mng.Add(ErrAccountDoesNotExist, httperr.Mapped{
		Status:  http.StatusUnauthorized,
		Code:    "invalid_credentials",
		Message: "Invalid username or password",
	})

	return mng
}
