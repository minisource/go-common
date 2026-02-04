package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// TenantConfig holds configuration for tenant middleware
type TenantConfig struct {
	// Enabled determines if tenant middleware is active
	Enabled bool
	// HeaderName is the header to look for tenant ID (e.g., "X-Tenant-ID")
	HeaderName string
	// ExtractFromSubdomain determines if tenant should be extracted from subdomain
	ExtractFromSubdomain bool
	// BaseDomain is the base domain for subdomain extraction (e.g., "example.com")
	BaseDomain string
	// SkipPaths are paths that don't require tenant context
	SkipPaths []string
	// AllowMissingTenant allows requests without tenant context to proceed
	AllowMissingTenant bool
	// ContextKey is the key to store tenant ID in context
	ContextKey string
	// ErrorHandler handles tenant resolution errors
	ErrorHandler fiber.ErrorHandler
	// TenantValidator validates if the tenant exists and is active
	// Returns true if valid, false otherwise
	TenantValidator func(tenantID string) bool
}

// DefaultTenantConfig returns default tenant configuration
func DefaultTenantConfig() TenantConfig {
	return TenantConfig{
		Enabled:            true,
		HeaderName:         "X-Tenant-ID",
		AllowMissingTenant: true,
		ContextKey:         "tenantId",
		SkipPaths:          []string{"/health", "/ready", "/metrics"},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid or missing tenant",
			})
		},
	}
}

// TenantMiddleware creates tenant context middleware for Fiber
// It extracts tenant ID from various sources in priority order:
// 1. JWT claims (if already authenticated)
// 2. Request header (X-Tenant-ID)
// 3. Subdomain (if configured)
func TenantMiddleware(config TenantConfig) fiber.Handler {
	// Set defaults
	if config.ContextKey == "" {
		config.ContextKey = "tenantId"
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-Tenant-ID"
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid or missing tenant",
			})
		}
	}

	return func(c *fiber.Ctx) error {
		// Check if middleware is disabled
		if !config.Enabled {
			return c.Next()
		}

		// Check if path should be skipped
		path := c.Path()
		for _, skipPath := range config.SkipPaths {
			if strings.HasPrefix(path, skipPath) || path == skipPath {
				return c.Next()
			}
		}

		var tenantID string

		// Priority 1: Check if tenant is already in context from JWT
		if existing := c.Locals("tenantId"); existing != nil {
			if tid, ok := existing.(string); ok && tid != "" {
				tenantID = tid
			}
		}

		// Priority 2: Check header
		if tenantID == "" {
			tenantID = c.Get(config.HeaderName)
		}

		// Priority 3: Check subdomain
		if tenantID == "" && config.ExtractFromSubdomain && config.BaseDomain != "" {
			tenantID = extractTenantFromSubdomain(c.Hostname(), config.BaseDomain)
		}

		// Priority 4: Check query parameter (for flexibility)
		if tenantID == "" {
			tenantID = c.Query("tenant_id")
		}

		// If no tenant ID found
		if tenantID == "" {
			if config.AllowMissingTenant {
				return c.Next()
			}
			return config.ErrorHandler(c, fiber.NewError(fiber.StatusBadRequest, "Tenant ID required"))
		}

		// Validate tenant if validator is provided
		if config.TenantValidator != nil {
			if !config.TenantValidator(tenantID) {
				return config.ErrorHandler(c, fiber.NewError(fiber.StatusBadRequest, "Invalid tenant"))
			}
		}

		// Store tenant ID in context
		c.Locals(config.ContextKey, tenantID)

		return c.Next()
	}
}

// extractTenantFromSubdomain extracts tenant from subdomain
// e.g., "tenant1.example.com" with baseDomain "example.com" returns "tenant1"
func extractTenantFromSubdomain(hostname, baseDomain string) string {
	// Remove port if present
	if idx := strings.Index(hostname, ":"); idx != -1 {
		hostname = hostname[:idx]
	}

	// Check if hostname ends with base domain
	if !strings.HasSuffix(hostname, baseDomain) {
		return ""
	}

	// Extract subdomain part
	subdomain := strings.TrimSuffix(hostname, "."+baseDomain)
	subdomain = strings.TrimSuffix(subdomain, baseDomain)

	// Skip if it's the root domain or www
	if subdomain == "" || subdomain == "www" || subdomain == "api" {
		return ""
	}

	return subdomain
}

// RequireTenant creates middleware that requires a valid tenant context
func RequireTenant() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID := c.Locals("tenantId")
		if tenantID == nil || tenantID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Tenant context required",
			})
		}
		return c.Next()
	}
}

// TenantScope creates middleware that enforces tenant access for specific operations
type TenantScopeConfig struct {
	// AllowSystemAccess allows requests without tenant context (for admin operations)
	AllowSystemAccess bool
	// SystemRoles are roles that can access without tenant context
	SystemRoles []string
}

// RequireTenantScope creates middleware that requires tenant context or system role
func RequireTenantScope(config TenantScopeConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID := c.Locals("tenantId")

		// If tenant is present, allow
		if tenantID != nil && tenantID != "" {
			return c.Next()
		}

		// Check for system access
		if config.AllowSystemAccess {
			roles, ok := c.Locals("roles").([]string)
			if ok {
				for _, role := range roles {
					for _, systemRole := range config.SystemRoles {
						if role == systemRole {
							return c.Next()
						}
					}
				}
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Tenant context required for this operation",
		})
	}
}

// GetTenantID is a helper to get tenant ID from Fiber context
func GetTenantID(c *fiber.Ctx) string {
	if tid := c.Locals("tenantId"); tid != nil {
		if s, ok := tid.(string); ok {
			return s
		}
	}
	return ""
}

// GetTenantIDPtr is a helper to get tenant ID as *string from Fiber context
func GetTenantIDPtr(c *fiber.Ctx) *string {
	tid := GetTenantID(c)
	if tid == "" {
		return nil
	}
	return &tid
}
