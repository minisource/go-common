package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Cors creates a middleware with custom CORS configuration.
// allowOrigins can be "*" or a comma-separated list of origins.
// Uses Origin request-header echo for proper CORS compliance with multiple origins.
func Cors(allowOrigins string) fiber.Handler {
	// Build origin lookup map for O(1) matching
	originMap := make(map[string]bool)
	for _, o := range strings.Split(allowOrigins, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			originMap[o] = true
		}
	}
	isWildcard := allowOrigins == "*"

	return func(c *fiber.Ctx) error {
		if isWildcard {
			c.Set("Access-Control-Allow-Origin", "*")
		} else {
			origin := c.Get("Origin")
			if originMap[origin] {
				c.Set("Access-Control-Allow-Origin", origin)
				c.Set("Access-Control-Allow-Credentials", "true")
			}
			// Non-matching origins: no Access-Control-Allow-Origin set → browser blocks the request
		}

		c.Set("Access-Control-Allow-Headers", "Authorization,Content-Type,Accept,Origin,X-Tenant-Id,X-Project-Id,X-Request-Id,Idempotency-Key")
		c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Set("Access-Control-Expose-Headers", "X-Request-Id")
		c.Set("Access-Control-Max-Age", "21600")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(204)
		}

		return c.Next()
	}
}
