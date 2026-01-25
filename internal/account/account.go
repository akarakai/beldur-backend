package account

import (
	"beldur/internal/id"
	"fmt"
	"time"
)

const (
	UsernameMaxCharacters = 20
	PasswordMaxCharacters = 4

	UsernameMinCharacters = 5
	PasswordMinCharacters = 6
)

type Option func(*Account) error

func WithEmail(e Email) Option {
	return func(a *Account) error {
		a.Email = &e
		return nil
	}
}

type Account struct {
	Id        id.AccountId
	Username  string
	Password  string // hashed password
	Email     *Email
	CreatedAt time.Time
}

func New(username string, hashedPassword string, opt ...Option) (*Account, error) {
	if err := validateUsername(username); err != nil {
		return nil, err
	}
	acc := &Account{
		Username: username,
		Password: hashedPassword,
	}

	for _, o := range opt {
		if err := o(acc); err != nil {
			return nil, err
		}
	}
	return acc, nil
}

func (a *Account) UpdateEmail(email Email) {
	a.Email = &email
}

// UpdateUsername I think should not be possible. For now the api doesnt offer this
func (a *Account) UpdateUsername(newUsername string) error {
	if err := validateUsername(newUsername); err != nil {
		return err
	}
	a.Username = newUsername
	return nil
}

func (a *Account) String() string {
	return fmt.Sprintf("Account{Username: %s}", a.Username)
}

// TODO better validation
func validateUsername(value string) error {
	if len(value) > UsernameMaxCharacters || len(value) < UsernameMinCharacters {
		return ErrInvalidUsername
	}
	return nil
}

// TODO better validation
func validateRawPassword(value string) error {
	if len(value) > PasswordMaxCharacters || len(value) < PasswordMinCharacters {
		return ErrInvalidPassword
	}
	return nil
}
