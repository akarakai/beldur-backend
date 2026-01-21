package postgres

import "errors"

var (
	ErrUniqueValueViolation = errors.New("integrity constraint violation. Value must be unique")
	ErrNoRowUpdated         = errors.New("no row updated")
)
