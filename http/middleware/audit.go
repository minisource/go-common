package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/go-common/audit"
)

// AuditConfig holds configuration for audit middleware
type AuditConfig struct {
	Logger         audit.Logger
	SkipPaths      []string
	SensitivePaths []string // Paths that should be audited
}

// DefaultAuditConfig returns default configuration
func DefaultAuditConfig(logger audit.Logger) *AuditConfig {
	return &AuditConfig{
		Logger: logger,
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/swagger",
		},
		SensitivePaths: []string{
			"/api/v1/auth/login",
			"/api/v1/auth/logout",
			"/api/v1/users",
			"/api/v1/roles",
			"/api/v1/permissions",
		},
	}
}

// AuditLogger creates audit logging middleware
func AuditLogger(config *AuditConfig) fiber.Handler {
	if config == nil {
		panic("audit config cannot be nil")
	}

	return func(c *fiber.Ctx) error {
		// Skip certain paths
		path := c.Path()
		for _, skip := range config.SkipPaths {
			if path == skip || len(path) >= len(skip) && path[:len(skip)] == skip {
				return c.Next()
			}
		}

		// Only audit sensitive paths
		shouldAudit := false
		for _, sensitive := range config.SensitivePaths {
			if path == sensitive || len(path) >= len(sensitive) && path[:len(sensitive)] == sensitive {
				shouldAudit = true
				break
			}
		}

		if !shouldAudit {
			return c.Next()
		}

		// Execute request
		err := c.Next()

		// Only log successful requests or specific status codes
		if c.Response().StatusCode() < 400 {
			// Get tenant and user from context
			tenantIDStr := c.Locals("tenantID")
			userIDStr := c.Locals("userID")

			var tenantID, userID uuid.UUID
			if tenantIDStr != nil {
				if tid, ok := tenantIDStr.(uuid.UUID); ok {
					tenantID = tid
				} else if tidStr, ok := tenantIDStr.(string); ok {
					tenantID, _ = uuid.Parse(tidStr)
				}
			}

			if userIDStr != nil {
				if uid, ok := userIDStr.(uuid.UUID); ok {
					userID = uid
				} else if uidStr, ok := userIDStr.(string); ok {
					userID, _ = uuid.Parse(uidStr)
				}
			}

			if tenantID != uuid.Nil {
				action := getActionFromMethod(c.Method())
				entityType := getEntityTypeFromPath(path)

				metadata := map[string]interface{}{
					"method":      c.Method(),
					"path":        path,
					"status_code": c.Response().StatusCode(),
				}

				_ = config.Logger.LogAction(
					c.Context(),
					tenantID,
					userID,
					action,
					entityType,
					nil,
					metadata,
				)
			}
		}

		return err
	}
}

func getActionFromMethod(method string) string {
	switch method {
	case "POST":
		return audit.ActionCreate
	case "PUT", "PATCH":
		return audit.ActionUpdate
	case "DELETE":
		return audit.ActionDelete
	case "GET":
		return audit.ActionView
	default:
		return "UNKNOWN"
	}
}

func getEntityTypeFromPath(path string) string {
	// Extract entity type from path
	// e.g., /api/v1/users -> USER
	if len(path) > 8 && path[:8] == "/api/v1/" {
		parts := path[8:]
		for i, char := range parts {
			if char == '/' {
				parts = parts[:i]
				break
			}
		}
		// Convert to uppercase
		entity := ""
		for _, char := range parts {
			if char >= 'a' && char <= 'z' {
				entity += string(char - 32)
			} else {
				entity += string(char)
			}
		}
		// Remove trailing 'S' for plurals
		if len(entity) > 0 && entity[len(entity)-1] == 'S' {
			entity = entity[:len(entity)-1]
		}
		return entity
	}
	return "UNKNOWN"
}
