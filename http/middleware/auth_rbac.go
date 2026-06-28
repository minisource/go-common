package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// ---------------------------------------------------------------------------
// RBAC helpers that extend the existing auth.go middleware.
// These assume that AuthMiddleware (or equivalent) has already stored
// userId, roles, permissions, etc. in Fiber Locals.
// ---------------------------------------------------------------------------

// GetEmailFromContext extracts email from Fiber context.
func GetEmailFromContext(c *fiber.Ctx) string {
	if email, ok := c.Locals("email").(string); ok {
		return email
	}
	return ""
}

// GetSessionIDFromContext extracts session ID from Fiber context.
func GetSessionIDFromContext(c *fiber.Ctx) string {
	if sid, ok := c.Locals("sessionId").(string); ok {
		return sid
	}
	return ""
}

// IsSuperAdmin checks if the current context has isSuperAdmin set.
func IsSuperAdmin(c *fiber.Ctx) bool {
	if super, ok := c.Locals("isSuperAdmin").(bool); ok {
		return super
	}
	return false
}

// IsAdmin determines if the current user should be treated as an admin.
// Admin = isSuperAdmin OR role in [admin, super_admin, system_admin] OR permission = notifier:admin.
func IsAdmin(c *fiber.Ctx) bool {
	if IsSuperAdmin(c) {
		return true
	}
	// Check roles
	for _, role := range GetRolesFromContext(c) {
		if role == "admin" || role == "super_admin" || role == "system_admin" {
			return true
		}
	}
	// Check permissions
	for _, perm := range GetPermissionsFromContext(c) {
		if perm == "notifier:admin" {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if the current user has any of the given permissions.
func HasAnyPermission(c *fiber.Ctx, permissions ...string) bool {
	for _, perm := range permissions {
		if HasPermission(c, perm) {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the current user has any of the given roles.
func HasAnyRole(c *fiber.Ctx, roles ...string) bool {
	for _, role := range roles {
		if HasRole(c, role) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Middleware creators (return fiber.Handler)
// ---------------------------------------------------------------------------

// RequireAuthenticated returns middleware that requires a valid authentication.
func RequireAuthenticated() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if GetUserIDFromContext(c) == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "AUTH_REQUIRED",
					"message": "Authentication required",
				},
			})
		}
		return c.Next()
	}
}

// RequireAdmin returns middleware that requires admin privileges.
func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !IsAdmin(c) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "FORBIDDEN",
					"message": "Admin permission is required",
				},
			})
		}
		return c.Next()
	}
}

// RequireRole returns middleware that requires a specific role.
func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !HasRole(c, role) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "FORBIDDEN",
					"message": "Required role: " + role,
				},
			})
		}
		return c.Next()
	}
}

// RequireAnyRole returns middleware that requires any of the given roles.
func RequireAnyRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !HasAnyRole(c, roles...) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "FORBIDDEN",
					"message": "Insufficient role permissions",
				},
			})
		}
		return c.Next()
	}
}

// RequirePermission returns middleware that requires a specific permission.
func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !HasPermission(c, permission) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "FORBIDDEN",
					"message": "Required permission: " + permission,
				},
			})
		}
		return c.Next()
	}
}

// RequireAnyPermission returns middleware that requires any of the given permissions.
func RequireAnyPermission(permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !HasAnyPermission(c, permissions...) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "FORBIDDEN",
					"message": "Insufficient permissions",
				},
			})
		}
		return c.Next()
	}
}

// RequireAdminOrPermission returns middleware that requires admin OR a specific permission.
func RequireAdminOrPermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if IsAdmin(c) || HasPermission(c, permission) {
			return c.Next()
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "FORBIDDEN",
				"message": "Admin or required permission: " + permission,
			},
		})
	}
}
