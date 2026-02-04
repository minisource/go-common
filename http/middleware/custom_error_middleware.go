package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func CustomErrorHandler(c *fiber.Ctx, err error) error {
	// Default to Internal Server Error
	code := fiber.StatusInternalServerError

	// Check if the error is of type *fiber.Error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Set the Content-Type header to application/json
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	// Return the error response as JSON
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
