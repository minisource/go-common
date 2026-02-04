package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// RequestValidationConfig defines configuration for request validation
type RequestValidationConfig struct {
	// MaxBodySize limits request body size in bytes (default: 4MB)
	MaxBodySize int64
	// MaxURILength limits URI length (default: 2048)
	MaxURILength int
	// MaxHeaderSize limits header size in bytes (default: 8KB)
	MaxHeaderSize int
	// AllowedMethods lists allowed HTTP methods
	AllowedMethods []string
	// BlockSuspiciousPatterns enables blocking of suspicious patterns in requests
	BlockSuspiciousPatterns bool
}

// DefaultRequestValidationConfig returns default validation configuration
func DefaultRequestValidationConfig() RequestValidationConfig {
	return RequestValidationConfig{
		MaxBodySize:             4 * 1024 * 1024, // 4MB
		MaxURILength:            2048,
		MaxHeaderSize:           8 * 1024, // 8KB
		AllowedMethods:          []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		BlockSuspiciousPatterns: true,
	}
}

var suspiciousPatterns = []string{
	"<script", "javascript:", "onerror=", "onload=",
	"../", "..\\", // Path traversal
	"union", "select", "drop", "insert", "update", "delete", // SQL injection basic patterns
	"eval(", "exec(", "system(", // Code injection
}

// RequestValidation middleware validates incoming requests
func RequestValidation(config RequestValidationConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check HTTP method
		if len(config.AllowedMethods) > 0 {
			methodAllowed := false
			for _, method := range config.AllowedMethods {
				if c.Method() == method {
					methodAllowed = true
					break
				}
			}
			if !methodAllowed {
				return c.Status(fiber.StatusMethodNotAllowed).JSON(fiber.Map{
					"error": "Method not allowed",
				})
			}
		}

		// Check URI length
		if config.MaxURILength > 0 && len(c.OriginalURL()) > config.MaxURILength {
			return c.Status(fiber.StatusRequestURITooLong).JSON(fiber.Map{
				"error": "Request URI too long",
			})
		}

		// Check Content-Length
		if config.MaxBodySize > 0 && c.Request().Header.ContentLength() > int(config.MaxBodySize) {
			return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"error": "Request body too large",
			})
		}

		// Check for suspicious patterns
		if config.BlockSuspiciousPatterns {
			uri := strings.ToLower(c.OriginalURL())
			for _, pattern := range suspiciousPatterns {
				if strings.Contains(uri, pattern) {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "Invalid request format",
					})
				}
			}

			// Check query parameters
			c.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
				lowerValue := strings.ToLower(string(value))
				for _, pattern := range suspiciousPatterns {
					if strings.Contains(lowerValue, pattern) {
						return
					}
				}
			})
		}

		return c.Next()
	}
}
