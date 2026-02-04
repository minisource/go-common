package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// Cors creates a middleware with custom CORS configuration
func Cors(allowOrigins string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Set CORS headers
        c.Set("Access-Control-Allow-Origin", allowOrigins)
        c.Set("Access-Control-Allow-Credentials", "true")
        c.Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
        c.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
        c.Set("Access-Control-Max-Age", "21600")
        c.Set("Content-Type", "application/json")

        // Handle preflight requests
        if c.Method() == "OPTIONS" {
            return c.SendStatus(204)
        }

        return c.Next()
    }
}

// Alternative using Fiber's built-in CORS middleware
func CorsWithConfig(allowOrigins string) fiber.Handler {
    return cors.New(cors.Config{
        AllowOrigins:     allowOrigins,
        AllowCredentials: true,
        AllowHeaders:     "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With",
        AllowMethods:     "POST, GET, OPTIONS, PUT, DELETE, UPDATE",
        MaxAge:           21600,
    })
}