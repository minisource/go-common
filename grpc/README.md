# Service Authentication

This package provides shared authentication middleware and gRPC interceptors for service-to-service OAuth2 authentication across all Minisource microservices.

## Overview

The authentication system supports two types of authentication:
1. **User JWT Authentication** - For direct API access by users (already in `http/middleware/auth.go`)
2. **Service Token Authentication** - For service-to-service communication using OAuth2 client credentials

## HTTP Middleware

### RemoteServiceAuthMiddleware

Validates service tokens by calling a remote auth service (OAuth2 introspection pattern).

```go
import (
    "github.com/minisource/go-common/http/middleware"
    "github.com/minisource/go-sdk/auth"
)

// Create auth client
authClient := auth.NewClient(auth.ClientConfig{
    BaseURL:      "http://auth-service:8080",
    ClientID:     "my-service",
    ClientSecret: "secret",
})

// Create adapter for the TokenValidator interface
validator := auth.NewHTTPClientAdapter(authClient)

// Create middleware
authMiddleware := middleware.RemoteServiceAuthMiddleware(middleware.RemoteServiceAuthConfig{
    TokenValidator: validator,
    Logger:         logger,
    CacheTTL:       5 * time.Minute,
    SkipPaths:      []string{"/health", "/ready"},
    RequiredScope:  "notifications:send",
    Enabled:        true,
})

// Use in Fiber
app.Use("/api/service", authMiddleware)
```

### RequireScope

Additional middleware to check for specific scopes:

```go
// Check for specific scope
app.Use("/api/service/send", middleware.RequireScope("notifications:send"))
```

## gRPC Interceptors

### UnaryAuthInterceptor / StreamAuthInterceptor

Validates service tokens in gRPC calls.

```go
import (
    commongrpc "github.com/minisource/go-common/grpc"
    "github.com/minisource/go-sdk/auth"
)

// Create auth client
authClient := auth.NewClient(auth.ClientConfig{
    BaseURL:      "http://auth-service:8080",
    ClientID:     "my-service",
    ClientSecret: "secret",
})

// Create gRPC adapter
validator := auth.NewGRPCClientAdapter(authClient)

// Define scope requirements for each method
scopeMap := map[string]string{
    "/myservice.v1.MyService/CreateItem": "items:create",
    "/myservice.v1.MyService/GetItem":    "items:read",
}

// Create interceptor config
cfg := commongrpc.AuthInterceptorConfig{
    TokenValidator: validator,
    Logger:         logger,
    CacheTTL:       5 * time.Minute,
    ScopeMap:       scopeMap,
    SkipMethods:    []string{"/grpc.health.v1.Health/Check"},
    Enabled:        true,
}

// Create gRPC server with interceptors
server := grpc.NewServer(
    grpc.UnaryInterceptor(commongrpc.UnaryAuthInterceptor(cfg)),
    grpc.StreamInterceptor(commongrpc.StreamAuthInterceptor(cfg)),
)
```

### Extracting Service Info from Context

In gRPC handlers, you can extract the validated service info:

```go
import commongrpc "github.com/minisource/go-common/grpc"

func (s *Server) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.CreateItemResponse, error) {
    clientID := commongrpc.GetServiceClientID(ctx)
    serviceName := commongrpc.GetServiceName(ctx)
    scopes := commongrpc.GetServiceScopes(ctx)
    tenantID := commongrpc.GetTenantID(ctx)
    
    // ... handler logic
}
```

## Token Caching

Both HTTP middleware and gRPC interceptors cache validated tokens to reduce calls to the auth service. The cache:
- Uses configurable TTL (default: 5 minutes)
- Automatically cleans up expired entries
- Can be cleared programmatically:

```go
// Clear HTTP token cache
middleware.ClearTokenCache()

// Clear gRPC token cache
commongrpc.ClearGRPCTokenCache()
```

## Scope Format

Scopes follow the format `resource:action`:
- `notifications:send` - Send notifications
- `notifications:read` - Read notifications
- `templates:*` - All template operations (wildcard)
- `*` - Full access (admin)

The scope checking supports:
- Exact match: `notifications:send` matches `notifications:send`
- Resource wildcard: `notifications:*` matches `notifications:send`, `notifications:read`, etc.
- Full wildcard: `*` matches any scope

## Integration with go-sdk/auth

The `go-sdk/auth` package provides adapters that implement the `TokenValidator` interface:

```go
// For HTTP middleware
httpAdapter := auth.NewHTTPClientAdapter(authClient)

// For gRPC interceptors
grpcAdapter := auth.NewGRPCClientAdapter(authClient)
```

These adapters call the auth service's `/api/v1/service/validate` endpoint to validate tokens.
