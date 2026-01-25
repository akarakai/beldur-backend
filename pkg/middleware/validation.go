package middleware

import (
	"beldur/pkg/httperr"
	"beldur/pkg/validation"

	"github.com/gofiber/fiber/v2"
)

// Validation is a middleware that  validates the request body and sets in c.Locals with the "body" key
func Validation[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req T
		if err := c.BodyParser(&req); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		if err := validation.Validate(req); err != nil {
			return httperr.ValidationFailed(c, err)
		}
		c.Locals("body", req)
		return c.Next()
	}
}
