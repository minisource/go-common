package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestIDConfig defines configuration for request ID middleware
type RequestIDConfig struct {
	// Header is the header name to look for request ID
	// Default: "X-Request-ID"
	Header string

	// Generator is a function to generate request ID
	// Default: uuid.New().String()
	Generator func() string

	// ContextKey is the key to store request ID in Fiber locals
	// Default: "request_id"
	ContextKey string
}

// DefaultRequestIDConfig returns default request ID configuration
func DefaultRequestIDConfig() RequestIDConfig {
	return RequestIDConfig{
		Header:     "X-Request-ID",
		Generator:  func() string { return uuid.New().String() },
		ContextKey: "request_id",
	}
}

// RequestID middleware adds a unique request ID to each request
// It checks for existing X-Request-ID header first, otherwise generates a new one
func RequestID(config ...RequestIDConfig) fiber.Handler {
	cfg := DefaultRequestIDConfig()
	if len(config) > 0 {
		cfg = config[0]
		// Set defaults for empty values
		if cfg.Header == "" {
			cfg.Header = "X-Request-ID"
		}
		if cfg.Generator == nil {
			cfg.Generator = func() string { return uuid.New().String() }
		}
		if cfg.ContextKey == "" {
			cfg.ContextKey = "request_id"
		}
	}

	return func(c *fiber.Ctx) error {
		// Check for existing request ID in header
		requestID := c.Get(cfg.Header)
		if requestID == "" {
			requestID = cfg.Generator()
		}

		// Store in locals for access in handlers
		c.Locals(cfg.ContextKey, requestID)

		// Set response header
		c.Set(cfg.Header, requestID)

		return c.Next()
	}
}

// GetRequestID retrieves the request ID from Fiber context locals
func GetRequestID(c *fiber.Ctx) string {
	if requestID, ok := c.Locals("request_id").(string); ok {
		return requestID
	}
	return ""
}
