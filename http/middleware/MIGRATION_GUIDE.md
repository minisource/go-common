# Middleware Migration Guide

This guide explains how to migrate from local middleware implementations to using go-common middleware.

## Available Middleware

| Middleware | File | Description |
|------------|------|-------------|
| `RequestID` | `request_id.go` | Adds unique request ID to each request |
| `ContentType` | `content_type.go` | Sets Content-Type header |
| `AuthMiddleware` | `auth.go` | JWT authentication (local validation) |
| `ServiceAuthMiddleware` | `auth.go` | Service-to-service JWT auth |
| `RemoteServiceAuthMiddleware` | `service_auth_remote.go` | Remote token validation via auth service |
| `TenantMiddleware` | `tenant.go` | Multi-tenant context extraction |
| `DefaultStructuredLogger` | `logger.go` | Request/response logging |
| `RateLimiter` | `limiter.go` | Rate limiting with Redis |
| `SecurityHeaders` | `security.go` | Security headers (XSS, HSTS, CSP, etc.) |
| `TracingMiddleware` | `tracing.go` | OpenTelemetry distributed tracing |
| `CORSMiddleware` | `cors.go` | Cross-origin resource sharing |
| `PrometheusMiddleware` | `prometheus.go` | Prometheus metrics |
| `AuditMiddleware` | `audit.go` | Audit logging |

## Migration Examples

### Before: Local RequestID Middleware

```go
// internal/middleware/middleware.go
package middleware

import (
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
)

func RequestID() fiber.Handler {
    return func(c *fiber.Ctx) error {
        requestID := c.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }
        c.Locals("request_id", requestID)
        c.Set("X-Request-ID", requestID)
        return c.Next()
    }
}
```

### After: Using go-common

```go
package router

import (
    commonMiddleware "github.com/minisource/go-common/http/middleware"
)

func (r *Router) Setup() *fiber.App {
    // Use go-common RequestID middleware
    r.app.Use(commonMiddleware.RequestID())
    
    // Or with custom config
    r.app.Use(commonMiddleware.RequestID(commonMiddleware.RequestIDConfig{
        Header:     "X-Request-ID",
        ContextKey: "request_id",
    }))
}
```

---

### Before: Local Auth Middleware

```go
// internal/middleware/auth.go
func AuthMiddleware(cfg *config.Config) fiber.Handler {
    authClient := auth.NewClient(auth.ClientConfig{
        BaseURL:      cfg.Auth.ServiceURL,
        ClientID:     cfg.Auth.ClientID,
        ClientSecret: cfg.Auth.ClientSecret,
    })
    
    return func(c *fiber.Ctx) error {
        authHeader := c.Get("Authorization")
        // ... validation logic
        claims, err := authClient.ValidateToken(ctx, token)
        // ... set locals
    }
}
```

### After: Using go-common RemoteServiceAuthMiddleware

```go
package router

import (
    commonMiddleware "github.com/minisource/go-common/http/middleware"
    "github.com/minisource/go-sdk/auth"
)

func (r *Router) Setup() *fiber.App {
    // Create auth client
    authClient := auth.NewClient(auth.ClientConfig{
        BaseURL:      r.config.Auth.ServiceURL,
        ClientID:     r.config.Auth.ClientID,
        ClientSecret: r.config.Auth.ClientSecret,
    })
    
    // Use go-common RemoteServiceAuthMiddleware with go-sdk auth validator
    r.app.Use(commonMiddleware.RemoteServiceAuthMiddleware(commonMiddleware.RemoteServiceAuthConfig{
        TokenValidator: &AuthTokenValidator{authClient: authClient},
        Enabled:        true,
        SkipPaths:      []string{"/health", "/ready"},
        Logger:         r.logger,
        CacheTTL:       5 * time.Minute,
    }))
}

// AuthTokenValidator adapts go-sdk auth client to go-common TokenValidator interface
type AuthTokenValidator struct {
    authClient *auth.Client
}

func (v *AuthTokenValidator) ValidateToken(ctx context.Context, token string) (*commonMiddleware.TokenValidationResult, error) {
    resp, err := v.authClient.ValidateToken(ctx, token)
    if err != nil {
        return nil, err
    }
    
    return &commonMiddleware.TokenValidationResult{
        Valid:       resp.Valid,
        ClientID:    resp.ClientID,
        ServiceName: resp.ServiceName,
        Scopes:      resp.Scopes,
        ExpiresAt:   time.Unix(resp.ExpiresAt, 0),
    }, nil
}
```

---

### Before: Local Tenant Middleware

```go
func TenantMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        tenantID := c.Get("X-Tenant-ID")
        if tenantID == "" {
            return response.BadRequest(c, "Tenant required")
        }
        c.Locals("tenant_id", tenantID)
        return c.Next()
    }
}
```

### After: Using go-common TenantMiddleware

```go
import commonMiddleware "github.com/minisource/go-common/http/middleware"

func (r *Router) Setup() *fiber.App {
    // Required tenant
    r.app.Use(commonMiddleware.TenantMiddleware(commonMiddleware.TenantConfig{
        Enabled:            true,
        HeaderName:         "X-Tenant-ID",
        AllowMissingTenant: false,
        ContextKey:         "tenant_id",
        SkipPaths:          []string{"/health", "/ready"},
    }))
    
    // Optional tenant
    r.app.Use(commonMiddleware.TenantMiddleware(commonMiddleware.TenantConfig{
        Enabled:            true,
        AllowMissingTenant: true,
    }))
}
```

---

### Before: Local Rate Limiter

```go
func RateLimitMiddleware(redisClient *redis.Client) fiber.Handler {
    return limiter.New(limiter.Config{
        Max:        100,
        Expiration: time.Minute,
        // ...
    })
}
```

### After: Using go-common RateLimiter

```go
import commonMiddleware "github.com/minisource/go-common/http/middleware"

func (r *Router) Setup() *fiber.App {
    r.app.Use(commonMiddleware.RateLimiter(commonMiddleware.RateLimiterConfig{
        Max:          100,
        Window:       time.Minute,
        RedisClient:  r.redisClient,
        SkipPaths:    []string{"/health"},
        KeyGenerator: func(c *fiber.Ctx) string {
            return c.IP()
        },
    }))
}
```

---

### Before: Local Logger Middleware

```go
func LoggingMiddleware(logger logging.Logger) fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := time.Now()
        err := c.Next()
        // log request
        return err
    }
}
```

### After: Using go-common Logger

```go
import commonMiddleware "github.com/minisource/go-common/http/middleware"

func (r *Router) Setup() *fiber.App {
    r.app.Use(commonMiddleware.DefaultStructuredLogger(&logging.LoggerConfig{
        Level:  "info",
        Logger: "zap",
    }))
}
```

---

## Standard Router Setup

Here's a complete example of a properly configured router using go-common:

```go
package router

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/recover"
    "github.com/minisource/go-common/logging"
    commonMiddleware "github.com/minisource/go-common/http/middleware"
    "github.com/minisource/go-sdk/auth"
    "github.com/myservice/config"
)

type Router struct {
    app        *fiber.App
    config     *config.Config
    logger     logging.Logger
    authClient *auth.Client
}

func NewRouter(cfg *config.Config, logger logging.Logger) *Router {
    app := fiber.New(fiber.Config{
        AppName:      "My Service",
        ErrorHandler: customErrorHandler,
    })
    
    authClient := auth.NewClient(auth.ClientConfig{
        BaseURL:      cfg.Auth.ServiceURL,
        ClientID:     cfg.Auth.ClientID,
        ClientSecret: cfg.Auth.ClientSecret,
    })
    
    return &Router{
        app:        app,
        config:     cfg,
        logger:     logger,
        authClient: authClient,
    }
}

func (r *Router) Setup() *fiber.App {
    // 1. Recovery middleware
    r.app.Use(recover.New())
    
    // 2. Request ID
    r.app.Use(commonMiddleware.RequestID())
    
    // 3. Security headers
    r.app.Use(commonMiddleware.SecurityHeaders(commonMiddleware.DefaultSecurityHeadersConfig()))
    
    // 4. CORS
    r.app.Use(commonMiddleware.CORSMiddleware(commonMiddleware.CORSConfig{
        AllowOrigins: "*",
        AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
        AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Tenant-ID,X-User-ID,X-Request-ID",
    }))
    
    // 5. Content type
    r.app.Use(commonMiddleware.ContentType())
    
    // 6. Logging
    r.app.Use(commonMiddleware.DefaultStructuredLogger(&logging.LoggerConfig{
        Level:  r.config.Log.Level,
        Logger: "zap",
    }))
    
    // 7. Prometheus metrics (optional)
    r.app.Use(commonMiddleware.PrometheusMiddleware(commonMiddleware.DefaultPrometheusConfig()))
    
    // Health routes (no auth required)
    r.app.Get("/health", r.healthHandler)
    r.app.Get("/ready", r.readyHandler)
    
    // API routes with auth
    api := r.app.Group("/api/v1")
    
    // Rate limiting
    api.Use(commonMiddleware.RateLimiter(commonMiddleware.RateLimiterConfig{
        Max:         r.config.RateLimit.RequestsPerMinute,
        Window:      time.Minute,
        SkipPaths:   []string{"/api/v1/public"},
    }))
    
    // Authentication
    api.Use(r.authMiddleware())
    
    // Tenant extraction
    api.Use(commonMiddleware.TenantMiddleware(commonMiddleware.TenantConfig{
        Enabled:            true,
        AllowMissingTenant: false,
    }))
    
    // Register route handlers
    r.registerRoutes(api)
    
    return r.app
}

func (r *Router) authMiddleware() fiber.Handler {
    return commonMiddleware.RemoteServiceAuthMiddleware(commonMiddleware.RemoteServiceAuthConfig{
        TokenValidator: &AuthTokenValidator{authClient: r.authClient},
        Enabled:        true,
        SkipPaths:      []string{"/health", "/ready", "/api/v1/public"},
        Logger:         r.logger,
        CacheTTL:       5 * time.Minute,
    })
}

// AuthTokenValidator implements middleware.TokenValidator
type AuthTokenValidator struct {
    authClient *auth.Client
}

func (v *AuthTokenValidator) ValidateToken(ctx context.Context, token string) (*commonMiddleware.TokenValidationResult, error) {
    resp, err := v.authClient.ValidateToken(ctx, token)
    if err != nil {
        return nil, err
    }
    return &commonMiddleware.TokenValidationResult{
        Valid:       resp.Valid,
        ClientID:    resp.ClientID,
        ServiceName: resp.ServiceName,
        Scopes:      resp.Scopes,
        ExpiresAt:   time.Unix(resp.ExpiresAt, 0),
    }, nil
}
```

## Checklist for Migration

- [ ] Replace local `RequestID` with `commonMiddleware.RequestID()`
- [ ] Replace local `Logger/Logging` with `commonMiddleware.DefaultStructuredLogger()`
- [ ] Replace local `Auth` with `commonMiddleware.RemoteServiceAuthMiddleware()` or `commonMiddleware.AuthMiddleware()`
- [ ] Replace local `Tenant` with `commonMiddleware.TenantMiddleware()`
- [ ] Replace local `RateLimit` with `commonMiddleware.RateLimiter()`
- [ ] Add `commonMiddleware.SecurityHeaders()` if missing
- [ ] Replace local `CORS` with `commonMiddleware.CORSMiddleware()` or Fiber's built-in
- [ ] Remove local middleware directory if all middleware migrated
- [ ] Update imports to use `github.com/minisource/go-common/http/middleware`
