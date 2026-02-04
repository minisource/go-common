package middleware

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/logging"
)

// TokenValidator is an interface for validating tokens via remote service
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (*TokenValidationResult, error)
}

// TokenValidationResult represents the result of token validation
type TokenValidationResult struct {
	Valid       bool
	ClientID    string
	ServiceName string
	TenantID    string
	UserID      string
	Scopes      []string
	ExpiresAt   time.Time
}

// RemoteServiceAuthConfig holds configuration for remote service auth middleware
type RemoteServiceAuthConfig struct {
	TokenValidator TokenValidator
	Logger         logging.Logger
	CacheTTL       time.Duration // TTL for token validation cache
	SkipPaths      []string      // Paths to skip authentication
	RequiredScope  string        // Required scope for this route group
	Enabled        bool          // Whether auth is enabled
}

// TokenValidationCache caches validated tokens
type TokenValidationCache struct {
	mu    sync.RWMutex
	cache map[string]*cachedTokenValidation
}

type cachedTokenValidation struct {
	result    *TokenValidationResult
	expiresAt time.Time
}

// Global token cache
var remoteTokenCache = &TokenValidationCache{
	cache: make(map[string]*cachedTokenValidation),
}

// RemoteServiceAuthMiddleware validates service JWT tokens using a remote auth service
func RemoteServiceAuthMiddleware(cfg RemoteServiceAuthConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if auth is disabled
		if !cfg.Enabled {
			return c.Next()
		}

		// Skip paths
		for _, path := range cfg.SkipPaths {
			if strings.HasPrefix(c.Path(), path) {
				return c.Next()
			}
		}

		// Get token from header
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing or invalid authorization header",
			})
		}
		token := authHeader[7:]

		// Check cache first
		if cached := remoteTokenCache.get(token); cached != nil {
			// Check scope if required
			if cfg.RequiredScope != "" && !hasScopeInList(cached.Scopes, cfg.RequiredScope) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Insufficient permissions",
				})
			}

			// Add service info to context
			c.Locals("serviceClientId", cached.ClientID)
			c.Locals("serviceName", cached.ServiceName)
			c.Locals("serviceScopes", cached.Scopes)
			c.Locals("tenantId", cached.TenantID)
			return c.Next()
		}

		// Validate token with auth service
		ctx := context.Background()
		validation, err := cfg.TokenValidator.ValidateToken(ctx, token)
		if err != nil {
			if cfg.Logger != nil {
				cfg.Logger.Error(logging.General, logging.Api, "Token validation failed", map[logging.ExtraKey]interface{}{
					"error": err.Error(),
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid service token",
			})
		}

		if !validation.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token is not valid",
			})
		}

		// Cache the validation result
		ttl := cfg.CacheTTL
		if ttl == 0 {
			ttl = 5 * time.Minute // Default 5 minutes
		}
		remoteTokenCache.set(token, validation, ttl)

		// Check scope if required
		if cfg.RequiredScope != "" && !hasScopeInList(validation.Scopes, cfg.RequiredScope) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":    "Insufficient permissions",
				"required": cfg.RequiredScope,
			})
		}

		// Add service info to context
		c.Locals("serviceClientId", validation.ClientID)
		c.Locals("serviceName", validation.ServiceName)
		c.Locals("serviceScopes", validation.Scopes)
		c.Locals("tenantId", validation.TenantID)

		return c.Next()
	}
}

// RequireScope creates a middleware that checks for a specific scope
func RequireScope(scope string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		scopes, ok := c.Locals("serviceScopes").([]string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "No scopes available",
			})
		}

		if !hasScopeInList(scopes, scope) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":    "Insufficient permissions",
				"required": scope,
			})
		}

		return c.Next()
	}
}

// hasScopeInList checks if the given scopes contain the required scope
func hasScopeInList(scopes []string, required string) bool {
	for _, scope := range scopes {
		// Check for wildcard scope
		if scope == "*" {
			return true
		}
		// Check exact match
		if scope == required {
			return true
		}
		// Check resource wildcard (e.g., "notifications:*" matches "notifications:send")
		parts := strings.Split(required, ":")
		if len(parts) == 2 {
			if scope == parts[0]+":*" {
				return true
			}
		}
	}
	return false
}

func (c *TokenValidationCache) get(token string) *TokenValidationResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, ok := c.cache[token]
	if !ok || time.Now().After(cached.expiresAt) {
		return nil
	}
	return cached.result
}

func (c *TokenValidationCache) set(token string, result *TokenValidationResult, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[token] = &cachedTokenValidation{
		result:    result,
		expiresAt: time.Now().Add(ttl),
	}

	// Clean up expired entries periodically
	if len(c.cache) > 1000 {
		c.cleanup()
	}
}

func (c *TokenValidationCache) cleanup() {
	now := time.Now()
	for k, v := range c.cache {
		if now.After(v.expiresAt) {
			delete(c.cache, k)
		}
	}
}

// ClearTokenCache clears the token validation cache (useful for testing)
func ClearTokenCache() {
	remoteTokenCache.mu.Lock()
	defer remoteTokenCache.mu.Unlock()
	remoteTokenCache.cache = make(map[string]*cachedTokenValidation)
}

// InvalidateToken removes a specific token from the cache
func InvalidateToken(token string) {
	remoteTokenCache.mu.Lock()
	defer remoteTokenCache.mu.Unlock()
	delete(remoteTokenCache.cache, token)
}
