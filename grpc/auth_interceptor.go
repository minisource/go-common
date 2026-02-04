package grpc

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/minisource/go-common/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

// AuthInterceptorConfig holds configuration for auth interceptors
type AuthInterceptorConfig struct {
	TokenValidator TokenValidator
	Logger         logging.Logger
	CacheTTL       time.Duration
	ScopeMap       map[string]string // Maps gRPC method to required scope
	SkipMethods    []string          // Methods that don't require authentication
	Enabled        bool
}

// Context keys for service info
type contextKey string

const (
	ServiceClientIDKey contextKey = "serviceClientId"
	ServiceNameKey     contextKey = "serviceName"
	ServiceScopesKey   contextKey = "serviceScopes"
	TenantIDKey        contextKey = "tenantId"
	UserIDKey          contextKey = "userId"
)

// grpcTokenCache caches validated tokens for gRPC
type grpcTokenCache struct {
	mu    sync.RWMutex
	cache map[string]*cachedValidation
}

type cachedValidation struct {
	result    *TokenValidationResult
	expiresAt time.Time
}

var tokenCache = &grpcTokenCache{
	cache: make(map[string]*cachedValidation),
}

// UnaryAuthInterceptor creates a gRPC unary interceptor for authentication
func UnaryAuthInterceptor(cfg AuthInterceptorConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Check if auth is disabled
		if !cfg.Enabled {
			return handler(ctx, req)
		}

		// Check if method should be skipped
		for _, method := range cfg.SkipMethods {
			if info.FullMethod == method {
				return handler(ctx, req)
			}
		}

		// Extract and validate token
		newCtx, err := validateGRPCToken(ctx, cfg)
		if err != nil {
			return nil, err
		}

		// Check for required scope based on method
		if cfg.ScopeMap != nil {
			requiredScope := cfg.ScopeMap[info.FullMethod]
			if requiredScope != "" {
				scopes, ok := newCtx.Value(ServiceScopesKey).([]string)
				if !ok || !HasScope(scopes, requiredScope) {
					return nil, status.Error(codes.PermissionDenied, "insufficient permissions")
				}
			}
		}

		return handler(newCtx, req)
	}
}

// StreamAuthInterceptor creates a gRPC stream interceptor for authentication
func StreamAuthInterceptor(cfg AuthInterceptorConfig) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Check if auth is disabled
		if !cfg.Enabled {
			return handler(srv, ss)
		}

		// Check if method should be skipped
		for _, method := range cfg.SkipMethods {
			if info.FullMethod == method {
				return handler(srv, ss)
			}
		}

		// Extract and validate token
		ctx := ss.Context()
		newCtx, err := validateGRPCToken(ctx, cfg)
		if err != nil {
			return err
		}

		// Check for required scope based on method
		if cfg.ScopeMap != nil {
			requiredScope := cfg.ScopeMap[info.FullMethod]
			if requiredScope != "" {
				scopes, ok := newCtx.Value(ServiceScopesKey).([]string)
				if !ok || !HasScope(scopes, requiredScope) {
					return status.Error(codes.PermissionDenied, "insufficient permissions")
				}
			}
		}

		// Wrap the stream with the new context
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          newCtx,
		}

		return handler(srv, wrapped)
	}
}

// validateGRPCToken validates the token from gRPC metadata
func validateGRPCToken(ctx context.Context, cfg AuthInterceptorConfig) (context.Context, error) {
	// Extract token from metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization header")
	}

	token := strings.TrimPrefix(authHeader[0], "Bearer ")
	if token == authHeader[0] {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	// Check cache first
	if cached := tokenCache.get(token); cached != nil {
		return addServiceInfoToContext(ctx, cached), nil
	}

	// Validate token with remote auth service
	validation, err := cfg.TokenValidator.ValidateToken(ctx, token)
	if err != nil {
		if cfg.Logger != nil {
			cfg.Logger.Error(logging.General, logging.Api, "Token validation failed", map[logging.ExtraKey]interface{}{
				"error": err.Error(),
			})
		}
		return nil, status.Error(codes.Unauthenticated, "token validation failed")
	}

	if !validation.Valid {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	// Cache the validation result
	ttl := cfg.CacheTTL
	if ttl == 0 {
		ttl = 5 * time.Minute
	}
	tokenCache.set(token, validation, ttl)

	return addServiceInfoToContext(ctx, validation), nil
}

func addServiceInfoToContext(ctx context.Context, validation *TokenValidationResult) context.Context {
	ctx = context.WithValue(ctx, ServiceClientIDKey, validation.ClientID)
	ctx = context.WithValue(ctx, ServiceNameKey, validation.ServiceName)
	ctx = context.WithValue(ctx, ServiceScopesKey, validation.Scopes)
	ctx = context.WithValue(ctx, TenantIDKey, validation.TenantID)
	if validation.UserID != "" {
		ctx = context.WithValue(ctx, UserIDKey, validation.UserID)
	}
	return ctx
}

// HasScope checks if the given scopes contain the required scope
func HasScope(scopes []string, required string) bool {
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

// wrappedServerStream wraps a grpc.ServerStream with a custom context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// Cache methods
func (c *grpcTokenCache) get(token string) *TokenValidationResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, ok := c.cache[token]
	if !ok || time.Now().After(cached.expiresAt) {
		return nil
	}
	return cached.result
}

func (c *grpcTokenCache) set(token string, result *TokenValidationResult, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[token] = &cachedValidation{
		result:    result,
		expiresAt: time.Now().Add(ttl),
	}

	// Clean up expired entries periodically
	if len(c.cache) > 1000 {
		c.cleanup()
	}
}

func (c *grpcTokenCache) cleanup() {
	now := time.Now()
	for k, v := range c.cache {
		if now.After(v.expiresAt) {
			delete(c.cache, k)
		}
	}
}

// ClearGRPCTokenCache clears the gRPC token validation cache
func ClearGRPCTokenCache() {
	tokenCache.mu.Lock()
	defer tokenCache.mu.Unlock()
	tokenCache.cache = make(map[string]*cachedValidation)
}

// Helper functions to extract values from context

// GetServiceClientID extracts service client ID from context
func GetServiceClientID(ctx context.Context) string {
	v, _ := ctx.Value(ServiceClientIDKey).(string)
	return v
}

// GetServiceName extracts service name from context
func GetServiceName(ctx context.Context) string {
	v, _ := ctx.Value(ServiceNameKey).(string)
	return v
}

// GetServiceScopes extracts service scopes from context
func GetServiceScopes(ctx context.Context) []string {
	v, _ := ctx.Value(ServiceScopesKey).([]string)
	return v
}

// GetTenantID extracts tenant ID from context
func GetTenantID(ctx context.Context) string {
	v, _ := ctx.Value(TenantIDKey).(string)
	return v
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) string {
	v, _ := ctx.Value(UserIDKey).(string)
	return v
}
