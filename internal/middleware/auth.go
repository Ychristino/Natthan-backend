package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/natthan/api/internal/core/token"
)

func Auth(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "token não fornecido",
			})
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := token.Parse(tokenStr, jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "token inválido ou expirado",
			})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("person_id", claims.PersonID)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("role").(string)
		if !ok || userRole != role {
			return fiber.NewError(fiber.StatusForbidden, "acesso negado")
		}
		return c.Next()
	}
}
