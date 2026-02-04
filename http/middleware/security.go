package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// SecurityHeadersConfig defines configuration for security headers middleware
type SecurityHeadersConfig struct {
	// XSSProtection enables X-XSS-Protection header
	XSSProtection bool
	// ContentTypeNosniff enables X-Content-Type-Options header
	ContentTypeNosniff bool
	// XFrameOptions sets X-Frame-Options header (DENY, SAMEORIGIN, ALLOW-FROM)
	XFrameOptions string
	// HSTSMaxAge sets Strict-Transport-Security max-age in seconds
	HSTSMaxAge int
	// HSTSIncludeSubdomains includes subdomains in HSTS
	HSTSIncludeSubdomains bool
	// ContentSecurityPolicy sets CSP header
	ContentSecurityPolicy string
	// ReferrerPolicy sets Referrer-Policy header
	ReferrerPolicy string
	// PermissionsPolicy sets Permissions-Policy header
	PermissionsPolicy string
}

// DefaultSecurityHeadersConfig returns default security headers configuration
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		XSSProtection:         true,
		ContentTypeNosniff:    true,
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            31536000, // 1 year
		HSTSIncludeSubdomains: true,
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		PermissionsPolicy:     "geolocation=(), microphone=(), camera=()",
	}
}

// SecurityHeaders middleware adds security headers to responses
func SecurityHeaders(config SecurityHeadersConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// X-XSS-Protection
		if config.XSSProtection {
			c.Set("X-XSS-Protection", "1; mode=block")
		}

		// X-Content-Type-Options
		if config.ContentTypeNosniff {
			c.Set("X-Content-Type-Options", "nosniff")
		}

		// X-Frame-Options
		if config.XFrameOptions != "" {
			c.Set("X-Frame-Options", config.XFrameOptions)
		}

		// Strict-Transport-Security (HSTS)
		if config.HSTSMaxAge > 0 {
			hstsValue := fmt.Sprintf("max-age=%d", config.HSTSMaxAge)
			if config.HSTSIncludeSubdomains {
				hstsValue += "; includeSubDomains"
			}
			c.Set("Strict-Transport-Security", hstsValue)
		}

		// Content-Security-Policy
		if config.ContentSecurityPolicy != "" {
			c.Set("Content-Security-Policy", config.ContentSecurityPolicy)
		}

		// Referrer-Policy
		if config.ReferrerPolicy != "" {
			c.Set("Referrer-Policy", config.ReferrerPolicy)
		}

		// Permissions-Policy
		if config.PermissionsPolicy != "" {
			c.Set("Permissions-Policy", config.PermissionsPolicy)
		}

		// Remove X-Powered-By header
		c.Set("X-Powered-By", "")

		return c.Next()
	}
}
