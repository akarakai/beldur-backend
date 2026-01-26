package validation

import "github.com/go-playground/validator/v10"

var v = validator.New()

func Validate(i any) error {
	return v.Struct(i)
}
