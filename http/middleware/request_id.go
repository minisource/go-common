package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// MaxRequestIDLength bounds accepted incoming request IDs. Longer values are
// rejected and replaced with a generated ID to prevent header/log bloat.
const MaxRequestIDLength = 128

// RequestIDConfig defines configuration for request ID middleware.
type RequestIDConfig struct {
	// Header is the header name to look for / write the request ID.
	// Default: "X-Request-ID"
	Header string

	// Generator produces a new request ID when none is supplied or the
	// supplied value is invalid. Default: uuid v4 string.
	Generator func() string

	// ContextKey is the Fiber locals key. Default: "request_id".
	ContextKey string

	// MaxLength bounds accepted incoming IDs (0 = unlimited).
	MaxLength int

	// SkipPaths are exact paths that bypass the middleware.
	SkipPaths []string
}

// DefaultRequestIDConfig returns the default request ID configuration.
func DefaultRequestIDConfig() RequestIDConfig {
	return RequestIDConfig{
		Header:     "X-Request-ID",
		Generator:  func() string { return uuid.New().String() },
		ContextKey: "request_id",
		MaxLength:  MaxRequestIDLength,
	}
}

// isValidRequestID reports whether an incoming request ID is safe to accept:
// non-empty, bounded in length, and made of printable ASCII characters only.
func isValidRequestID(id string, max int) bool {
	if id == "" {
		return false
	}
	if max > 0 && len(id) > max {
		return false
	}
	for _, r := range id {
		if r < 0x21 || r > 0x7E {
			return false // control characters or non-ASCII
		}
	}
	return true
}

// RequestID middleware adds a unique request ID to each request. It accepts a
// valid incoming X-Request-ID, otherwise generates one. The ID is stored in
// Fiber locals, echoed in the response header, and set on the request header
// so outbound calls to downstream services propagate it.
func RequestID(config ...RequestIDConfig) fiber.Handler {
	cfg := DefaultRequestIDConfig()
	if len(config) > 0 {
		if config[0].Header != "" {
			cfg.Header = config[0].Header
		}
		if config[0].Generator != nil {
			cfg.Generator = config[0].Generator
		}
		if config[0].ContextKey != "" {
			cfg.ContextKey = config[0].ContextKey
		}
		if config[0].MaxLength != 0 {
			cfg.MaxLength = config[0].MaxLength
		}
		if config[0].SkipPaths != nil {
			cfg.SkipPaths = config[0].SkipPaths
		}
	}

	return func(c *fiber.Ctx) error {
		for _, p := range cfg.SkipPaths {
			if c.Path() == p {
				return c.Next()
			}
		}

		requestID := c.Get(cfg.Header)
		if !isValidRequestID(requestID, cfg.MaxLength) {
			requestID = cfg.Generator()
		}

		// Store in locals for access in handlers.
		c.Locals(cfg.ContextKey, requestID)

		// Echo back to the client and set on the request headers so any
		// outbound call (e.g. gateway proxy) forwards the same ID.
		c.Set(cfg.Header, requestID)
		c.Request().Header.Set(cfg.Header, requestID)

		return c.Next()
	}
}

// GetRequestID retrieves the request ID from Fiber context locals. It falls
// back to the "requestId" key for backward compatibility with older services.
func GetRequestID(c *fiber.Ctx) string {
	if id, ok := c.Locals("request_id").(string); ok && id != "" {
		return id
	}
	if id, ok := c.Locals("requestId").(string); ok && id != "" {
		return id
	}
	return strings.TrimSpace(c.Get("X-Request-ID"))
}
