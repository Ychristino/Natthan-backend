package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "erro interno do servidor"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"status":    code,
		"error":     "error",
		"message":   message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"path":      c.Path(),
	})
}
