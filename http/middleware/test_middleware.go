package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func TestMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := c.Get("x-api-key")
		if apiKey == "1" {
			return c.Next()
		}
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"result": "Api key is required",
		})
	}
}