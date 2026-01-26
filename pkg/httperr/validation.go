package httperr

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func ValidationFailed(c *fiber.Ctx, err error) error {
	var fields []string

	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		for _, fe := range verrs {
			fields = append(fields, fe.Field())
		}
	}

	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"code":           "invalid_request",
		"message":        "validation failed",
		"invalid_fields": fields,
		"timestamp":      time.Now().UTC(),
	})
}
