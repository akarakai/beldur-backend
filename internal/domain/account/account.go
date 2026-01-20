package account

import (
	"beldur/internal/domain/account/email"
	"time"
)

const (
	UsernameMaxCharacters = 10
	PasswordMaxCharacters = 10
)

type AccountOpt func(*Account) error

func WithEmail(emailValue string) AccountOpt {
	return func(a *Account) error {
		e, err := email.New(emailValue)
		if err != nil {
			return err
		}
		a.Email = e
		return nil
	}
}

type Account struct {
	Id        int
	Username  string
	Password  string // hashed password
	Email     email.Email
	CreatedAt time.Time
}

func New(username string, hashedPassword string, opt ...AccountOpt) (*Account, error) {
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

func (a *Account) ChangeUsername(newUsername string) error {
	if err := validateUsername(newUsername); err != nil {
		return err
	}
	a.Username = newUsername
	return nil
}

// TODO better validation
func validateUsername(value string) error {
	if len(value) > UsernameMaxCharacters {
		return ErrInvalidUsername
	}
	return nil
}

// TODO better validation
func validateRawPassword(value string) error {
	if len(value) > PasswordMaxCharacters {
		return ErrInvalidPassword
	}
	return nil
}
