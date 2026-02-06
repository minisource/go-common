package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// ContentTypeConfig defines configuration for content type middleware
type ContentTypeConfig struct {
	// ContentType is the content type to set
	// Default: "application/json"
	ContentType string

	// Charset is the charset to append to content type
	// Default: "utf-8"
	Charset string

	// SkipPaths are paths to skip setting content type
	SkipPaths []string
}

// DefaultContentTypeConfig returns default content type configuration
func DefaultContentTypeConfig() ContentTypeConfig {
	return ContentTypeConfig{
		ContentType: "application/json",
		Charset:     "utf-8",
		SkipPaths:   []string{},
	}
}

// ContentType middleware sets the Content-Type header for responses
func ContentType(config ...ContentTypeConfig) fiber.Handler {
	cfg := DefaultContentTypeConfig()
	if len(config) > 0 {
		cfg = config[0]
		// Set defaults for empty values
		if cfg.ContentType == "" {
			cfg.ContentType = "application/json"
		}
		if cfg.Charset == "" {
			cfg.Charset = "utf-8"
		}
	}

	// Build content type string
	contentType := cfg.ContentType
	if cfg.Charset != "" {
		contentType = contentType + "; charset=" + cfg.Charset
	}

	return func(c *fiber.Ctx) error {
		// Check if path should be skipped
		for _, path := range cfg.SkipPaths {
			if c.Path() == path {
				return c.Next()
			}
		}

		// Set content type header
		c.Set("Content-Type", contentType)

		return c.Next()
	}
}

// JSONContentType is a shorthand for ContentType with application/json
func JSONContentType() fiber.Handler {
	return ContentType(ContentTypeConfig{
		ContentType: "application/json",
		Charset:     "utf-8",
	})
}

// XMLContentType is a shorthand for ContentType with application/xml
func XMLContentType() fiber.Handler {
	return ContentType(ContentTypeConfig{
		ContentType: "application/xml",
		Charset:     "utf-8",
	})
}

// HTMLContentType is a shorthand for ContentType with text/html
func HTMLContentType() fiber.Handler {
	return ContentType(ContentTypeConfig{
		ContentType: "text/html",
		Charset:     "utf-8",
	})
}
