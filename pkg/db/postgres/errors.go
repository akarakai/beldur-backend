package postgres

import "errors"

var (
	ErrUniqueValueViolation = errors.New("integrity constraint violation. Value must be unique")
	ErrNoRowUpdated         = errors.New("no row updated")
	ErrNoRowFound           = errors.New("no row found")
)
