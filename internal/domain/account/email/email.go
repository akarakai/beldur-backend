package email

import (
	"errors"
	"net/mail"
	"strings"
)

var (
	ErrInvalidEmailFormat = errors.New("invalid email format")
)

type Email struct {
	value string
}

func New(email string) (Email, error) {
	email = strings.TrimSpace(email)

	_, err := mail.ParseAddress(email)
	if err != nil {
		return Email{}, ErrInvalidEmailFormat
	}
	email = strings.ToLower(email)
	return Email{value: email}, nil
}

func (e Email) String() string {
	return e.value
}

// I think it cannot be null because I validate in new....
func (e Email) IsNull() bool {
	return e.value == ""
}
