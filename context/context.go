package context

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Context keys for storing values
type contextKey string

const (
	keyUserID      contextKey = "user_id"
	keyTenantID    contextKey = "tenant_id"
	keySessionID   contextKey = "session_id"
	keyTraceID     contextKey = "trace_id"
	keySpanID      contextKey = "span_id"
	keyRequestID   contextKey = "request_id"
	keyRoles       contextKey = "roles"
	keyPermissions contextKey = "permissions"
	keyLanguage    contextKey = "language"
	keyClientIP    contextKey = "client_ip"
	keyUserAgent   contextKey = "user_agent"
)

// RequestContext holds all request-scoped data
type RequestContext struct {
	UserID      uuid.UUID
	TenantID    uuid.UUID
	SessionID   string
	TraceID     string
	SpanID      string
	RequestID   string
	Roles       []string
	Permissions []string
	Language    string
	ClientIP    string
	UserAgent   string
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, keyUserID, userID)
}

// GetUserID retrieves user ID from context
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(keyUserID).(uuid.UUID)
	return id, ok
}

// MustGetUserID retrieves user ID or panics
func MustGetUserID(ctx context.Context) uuid.UUID {
	id, ok := GetUserID(ctx)
	if !ok {
		panic("user ID not found in context")
	}
	return id
}

// WithTenantID adds tenant ID to context
func WithTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, keyTenantID, tenantID)
}

// GetTenantID retrieves tenant ID from context
func GetTenantID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(keyTenantID).(uuid.UUID)
	return id, ok
}

// MustGetTenantID retrieves tenant ID or panics
func MustGetTenantID(ctx context.Context) uuid.UUID {
	id, ok := GetTenantID(ctx)
	if !ok {
		panic("tenant ID not found in context")
	}
	return id
}

// WithSessionID adds session ID to context
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, keySessionID, sessionID)
}

// GetSessionID retrieves session ID from context
func GetSessionID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(keySessionID).(string)
	return id, ok
}

// WithTraceID adds trace ID to context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, keyTraceID, traceID)
}

// GetTraceID retrieves trace ID from context
func GetTraceID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(keyTraceID).(string)
	return id, ok
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, keyRequestID, requestID)
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(keyRequestID).(string)
	return id, ok
}

// WithRoles adds roles to context
func WithRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, keyRoles, roles)
}

// GetRoles retrieves roles from context
func GetRoles(ctx context.Context) []string {
	roles, _ := ctx.Value(keyRoles).([]string)
	return roles
}

// HasRole checks if context has a specific role
func HasRole(ctx context.Context, role string) bool {
	roles := GetRoles(ctx)
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// WithPermissions adds permissions to context
func WithPermissions(ctx context.Context, permissions []string) context.Context {
	return context.WithValue(ctx, keyPermissions, permissions)
}

// GetPermissions retrieves permissions from context
func GetPermissions(ctx context.Context) []string {
	perms, _ := ctx.Value(keyPermissions).([]string)
	return perms
}

// HasPermission checks if context has a specific permission
func HasPermission(ctx context.Context, permission string) bool {
	perms := GetPermissions(ctx)
	for _, p := range perms {
		if p == permission {
			return true
		}
	}
	return false
}

// WithLanguage adds language to context
func WithLanguage(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, keyLanguage, lang)
}

// GetLanguage retrieves language from context
func GetLanguage(ctx context.Context) string {
	lang, _ := ctx.Value(keyLanguage).(string)
	if lang == "" {
		return "en"
	}
	return lang
}

// WithClientIP adds client IP to context
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, keyClientIP, ip)
}

// GetClientIP retrieves client IP from context
func GetClientIP(ctx context.Context) string {
	ip, _ := ctx.Value(keyClientIP).(string)
	return ip
}

// WithRequestContext adds all request context values
func WithRequestContext(ctx context.Context, rc *RequestContext) context.Context {
	if rc.UserID != uuid.Nil {
		ctx = WithUserID(ctx, rc.UserID)
	}
	if rc.TenantID != uuid.Nil {
		ctx = WithTenantID(ctx, rc.TenantID)
	}
	if rc.SessionID != "" {
		ctx = WithSessionID(ctx, rc.SessionID)
	}
	if rc.TraceID != "" {
		ctx = WithTraceID(ctx, rc.TraceID)
	}
	if rc.RequestID != "" {
		ctx = WithRequestID(ctx, rc.RequestID)
	}
	if len(rc.Roles) > 0 {
		ctx = WithRoles(ctx, rc.Roles)
	}
	if len(rc.Permissions) > 0 {
		ctx = WithPermissions(ctx, rc.Permissions)
	}
	if rc.Language != "" {
		ctx = WithLanguage(ctx, rc.Language)
	}
	if rc.ClientIP != "" {
		ctx = WithClientIP(ctx, rc.ClientIP)
	}
	return ctx
}

// GetRequestContext extracts all request context values
func GetRequestContext(ctx context.Context) *RequestContext {
	rc := &RequestContext{}
	rc.UserID, _ = GetUserID(ctx)
	rc.TenantID, _ = GetTenantID(ctx)
	rc.SessionID, _ = GetSessionID(ctx)
	rc.TraceID, _ = GetTraceID(ctx)
	rc.RequestID, _ = GetRequestID(ctx)
	rc.Roles = GetRoles(ctx)
	rc.Permissions = GetPermissions(ctx)
	rc.Language = GetLanguage(ctx)
	rc.ClientIP = GetClientIP(ctx)
	return rc
}

// ============================================
// Fiber Context Helpers
// ============================================

// FromFiber extracts context from Fiber and adds request metadata
func FromFiber(c *fiber.Ctx) context.Context {
	ctx := c.UserContext()

	// Add trace ID if present
	if traceID, ok := c.Locals("traceId").(string); ok {
		ctx = WithTraceID(ctx, traceID)
	}

	// Add request ID
	requestID := c.Get("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}
	ctx = WithRequestID(ctx, requestID)

	// Add client IP
	ctx = WithClientIP(ctx, c.IP())

	// Add language
	lang := c.Get("Accept-Language")
	if lang == "" {
		lang = c.Query("lang", "en")
	}
	if len(lang) >= 2 {
		lang = lang[:2]
	}
	ctx = WithLanguage(ctx, lang)

	return ctx
}

// SetToFiber stores context in Fiber locals
func SetToFiber(c *fiber.Ctx, ctx context.Context) {
	c.SetUserContext(ctx)

	// Also store in locals for easy access in handlers
	if userID, ok := GetUserID(ctx); ok {
		c.Locals("userId", userID.String())
	}
	if tenantID, ok := GetTenantID(ctx); ok {
		c.Locals("tenantId", tenantID.String())
	}
	if traceID, ok := GetTraceID(ctx); ok {
		c.Locals("traceId", traceID)
	}
	if requestID, ok := GetRequestID(ctx); ok {
		c.Locals("requestId", requestID)
	}
}

// GetUserIDFromFiber gets user ID from Fiber context
func GetUserIDFromFiber(c *fiber.Ctx) (uuid.UUID, bool) {
	if idStr, ok := c.Locals("userId").(string); ok {
		if id, err := uuid.Parse(idStr); err == nil {
			return id, true
		}
	}
	return GetUserID(c.UserContext())
}

// GetTenantIDFromFiber gets tenant ID from Fiber context
func GetTenantIDFromFiber(c *fiber.Ctx) (uuid.UUID, bool) {
	if idStr, ok := c.Locals("tenantId").(string); ok {
		if id, err := uuid.Parse(idStr); err == nil {
			return id, true
		}
	}
	return GetTenantID(c.UserContext())
}
