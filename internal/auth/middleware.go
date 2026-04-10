package auth

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"boilerplate/internal/domain"
)

// Middleware validates the Authorization header using the Clean AuthUseCase
func Middleware(uc domain.AuthUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing Authorization header",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid Authorization format",
			})
		}

		token := parts[1]

		// Use Domain layer to validate the token, remaining agnostic to PocketBase.
		user, err := uc.ValidateToken(c.Context(), token)
		if err != nil || user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Store domain.User in context
		c.Locals("user", user)

		return c.Next()
	}
}
