package middleware

import (
	"beldur/pkg/auth"

	"github.com/gofiber/fiber/v2"
)

const principalKey = "principal"

func Auth(verifier auth.TokenVerifier) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies("jwt")
		if token == "" {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		verified, err := verifier.Verify(c.Context(), token)
		if err != nil {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		c.Locals(principalKey, auth.Principal{
			AccountID: verified.Subject,
			PlayerID:  verified.PlayerId,
		})

		return c.Next()
	}
}

func PrincipalFromCtx(c *fiber.Ctx) (auth.Principal, bool) {
	v := c.Locals(principalKey)
	p, ok := v.(auth.Principal)
	return p, ok
}
