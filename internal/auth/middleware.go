package auth

import (
	"github.com/gofiber/fiber/v2"
)

const principalKey = "principal"

func Middleware(verifier TokenVerifier) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		token := c.Cookies("jwt")
		if token == "" {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		verified, err := verifier.Verify(c.Context(), token)
		if err != nil {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		c.Locals(principalKey, Principal{
			AccountID: verified.Subject,
			PlayerID:  verified.PlayerId,
		})

		return c.Next()
	}
}

func PrincipalFromCtx(c *fiber.Ctx) (Principal, bool) {
	v := c.Locals(principalKey)
	p, ok := v.(Principal)
	return p, ok
}
