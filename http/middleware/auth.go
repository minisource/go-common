package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// AuthConfig holds configuration for auth middleware
type AuthConfig struct {
	// Enabled determines if auth is required
	Enabled bool
	// Secret is the JWT signing key
	Secret string
	// SkipPaths are paths that don't require authentication
	SkipPaths []string
	// TokenLookup defines how to extract token
	// Format: "header:Authorization" or "query:token" or "cookie:jwt"
	TokenLookup string
	// AuthScheme is the authorization scheme (default: Bearer)
	AuthScheme string
	// ContextKey is the key to store claims in context
	ContextKey string
	// ErrorHandler handles authentication errors
	ErrorHandler fiber.ErrorHandler
	// SuccessHandler is called after successful auth
	SuccessHandler fiber.Handler
	// Validator is custom token validation function
	Validator func(token string) (*TokenClaims, error)
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID      string   `json:"userId"`
	SessionID   string   `json:"sessionId"`
	Email       string   `json:"email"`
	TenantID    string   `json:"tenantId,omitempty"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	TokenType   string   `json:"tokenType"`
	jwt.RegisteredClaims
}

// ServiceTokenClaims represents service-to-service JWT claims
type ServiceTokenClaims struct {
	ClientID    string   `json:"clientId"`
	ServiceName string   `json:"serviceName"`
	TenantID    string   `json:"tenantId,omitempty"`
	Scopes      []string `json:"scopes"`
	TokenType   string   `json:"tokenType"`
	jwt.RegisteredClaims
}

// DefaultAuthConfig returns default auth configuration
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		Enabled:     true,
		TokenLookup: "header:Authorization",
		AuthScheme:  "Bearer",
		ContextKey:  "user",
		SkipPaths:   []string{"/health", "/ready", "/api/v1/auth/login", "/api/v1/auth/register"},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		},
	}
}

// AuthMiddleware creates JWT authentication middleware for Fiber
func AuthMiddleware(config AuthConfig) fiber.Handler {
	// Set defaults
	if config.TokenLookup == "" {
		config.TokenLookup = "header:Authorization"
	}
	if config.AuthScheme == "" {
		config.AuthScheme = "Bearer"
	}
	if config.ContextKey == "" {
		config.ContextKey = "user"
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}
	}

	return func(c *fiber.Ctx) error {
		// Check if auth is disabled
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

		// Extract token
		token := extractTokenFromRequest(c, config.TokenLookup, config.AuthScheme)
		if token == "" {
			return config.ErrorHandler(c, fiber.NewError(fiber.StatusUnauthorized, "No token provided"))
		}

		// Validate token
		var claims *TokenClaims
		var err error

		if config.Validator != nil {
			claims, err = config.Validator(token)
		} else {
			claims, err = validateToken(token, config.Secret)
		}

		if err != nil {
			return config.ErrorHandler(c, err)
		}

		// Store claims in context
		c.Locals(config.ContextKey, claims)
		c.Locals("userId", claims.UserID)
		c.Locals("sessionId", claims.SessionID)
		c.Locals("email", claims.Email)
		c.Locals("tenantId", claims.TenantID)
		c.Locals("roles", claims.Roles)
		c.Locals("permissions", claims.Permissions)

		// Call success handler if provided
		if config.SuccessHandler != nil {
			return config.SuccessHandler(c)
		}

		return c.Next()
	}
}

// ServiceAuthMiddleware creates service-to-service authentication middleware
func ServiceAuthMiddleware(config AuthConfig) fiber.Handler {
	if config.TokenLookup == "" {
		config.TokenLookup = "header:Authorization"
	}
	if config.AuthScheme == "" {
		config.AuthScheme = "Bearer"
	}
	if config.ContextKey == "" {
		config.ContextKey = "service"
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}
	}

	return func(c *fiber.Ctx) error {
		if !config.Enabled {
			return c.Next()
		}

		// Check if path should be skipped
		path := c.Path()
		for _, skipPath := range config.SkipPaths {
			if strings.HasPrefix(path, skipPath) {
				return c.Next()
			}
		}

		token := extractTokenFromRequest(c, config.TokenLookup, config.AuthScheme)
		if token == "" {
			return config.ErrorHandler(c, fiber.NewError(fiber.StatusUnauthorized, "No token provided"))
		}

		claims, err := validateServiceToken(token, config.Secret)
		if err != nil {
			return config.ErrorHandler(c, err)
		}

		c.Locals(config.ContextKey, claims)
		c.Locals("clientId", claims.ClientID)
		c.Locals("serviceName", claims.ServiceName)
		c.Locals("tenantId", claims.TenantID)
		c.Locals("scopes", claims.Scopes)

		return c.Next()
	}
}

// RequireRoles creates middleware that requires specific roles
func RequireRoles(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRoles, ok := c.Locals("roles").([]string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		for _, required := range roles {
			for _, userRole := range userRoles {
				if userRole == required {
					return c.Next()
				}
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions",
		})
	}
}

// RequirePermissions creates middleware that requires specific permissions
func RequirePermissions(permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userPerms, ok := c.Locals("permissions").([]string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		for _, required := range permissions {
			found := false
			for _, userPerm := range userPerms {
				if userPerm == required {
					found = true
					break
				}
			}
			if !found {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Insufficient permissions",
				})
			}
		}

		return c.Next()
	}
}

// RequireScopes creates middleware that requires specific scopes (for service auth)
func RequireScopes(scopes ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientScopes, ok := c.Locals("scopes").([]string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		for _, required := range scopes {
			found := false
			for _, scope := range clientScopes {
				if scope == required || scope == "*" {
					found = true
					break
				}
			}
			if !found {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Insufficient scopes",
				})
			}
		}

		return c.Next()
	}
}

// OptionalAuth creates middleware that sets user context if token is present
// but doesn't require authentication
func OptionalAuth(config AuthConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractTokenFromRequest(c, config.TokenLookup, config.AuthScheme)
		if token == "" {
			return c.Next()
		}

		var claims *TokenClaims
		var err error

		if config.Validator != nil {
			claims, err = config.Validator(token)
		} else {
			claims, err = validateToken(token, config.Secret)
		}

		if err == nil && claims != nil {
			c.Locals(config.ContextKey, claims)
			c.Locals("userId", claims.UserID)
			c.Locals("sessionId", claims.SessionID)
			c.Locals("email", claims.Email)
			c.Locals("roles", claims.Roles)
			c.Locals("permissions", claims.Permissions)
		}

		return c.Next()
	}
}

// Helper functions

func extractTokenFromRequest(c *fiber.Ctx, lookup, scheme string) string {
	parts := strings.Split(lookup, ":")
	if len(parts) != 2 {
		return ""
	}

	switch parts[0] {
	case "header":
		return extractFromHeader(c, parts[1], scheme)
	case "query":
		return c.Query(parts[1])
	case "cookie":
		return c.Cookies(parts[1])
	}

	return ""
}

func extractFromHeader(c *fiber.Ctx, header, scheme string) string {
	auth := c.Get(header)
	if auth == "" {
		return ""
	}

	if scheme != "" {
		schemeLen := len(scheme)
		if len(auth) > schemeLen && strings.EqualFold(auth[:schemeLen], scheme) {
			return strings.TrimSpace(auth[schemeLen:])
		}
		return ""
	}

	return auth
}

func validateToken(tokenString, secret string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token claims")
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Token expired")
	}

	return claims, nil
}

func validateServiceToken(tokenString, secret string) (*ServiceTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ServiceTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}

	claims, ok := token.Claims.(*ServiceTokenClaims)
	if !ok || !token.Valid {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token claims")
	}

	if claims.TokenType != "service" {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token type")
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Token expired")
	}

	return claims, nil
}

// GetUserIDFromContext extracts user ID from fiber context
func GetUserIDFromContext(c *fiber.Ctx) string {
	userID, ok := c.Locals("userId").(string)
	if !ok {
		return ""
	}
	return userID
}

// GetRolesFromContext extracts roles from fiber context
func GetRolesFromContext(c *fiber.Ctx) []string {
	roles, ok := c.Locals("roles").([]string)
	if !ok {
		return nil
	}
	return roles
}

// GetPermissionsFromContext extracts permissions from fiber context
func GetPermissionsFromContext(c *fiber.Ctx) []string {
	perms, ok := c.Locals("permissions").([]string)
	if !ok {
		return nil
	}
	return perms
}

// GetClaimsFromContext extracts token claims from fiber context
func GetClaimsFromContext(c *fiber.Ctx, key string) *TokenClaims {
	if key == "" {
		key = "user"
	}
	claims, ok := c.Locals(key).(*TokenClaims)
	if !ok {
		return nil
	}
	return claims
}

// GetServiceClaimsFromContext extracts service token claims from fiber context
func GetServiceClaimsFromContext(c *fiber.Ctx) *ServiceTokenClaims {
	claims, ok := c.Locals("service").(*ServiceTokenClaims)
	if !ok {
		return nil
	}
	return claims
}

// HasRole checks if user has specific role
func HasRole(c *fiber.Ctx, role string) bool {
	roles := GetRolesFromContext(c)
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasPermission checks if user has specific permission
func HasPermission(c *fiber.Ctx, permission string) bool {
	perms := GetPermissionsFromContext(c)
	for _, p := range perms {
		if p == permission {
			return true
		}
	}
	return false
}

// HasScope checks if service has specific scope
func HasScope(c *fiber.Ctx, scope string) bool {
	scopes, ok := c.Locals("scopes").([]string)
	if !ok {
		return false
	}
	for _, s := range scopes {
		if s == scope || s == "*" {
			return true
		}
	}
	return false
}
