# Minisource Go Common Library

Shared Go library providing common utilities, middleware, and helpers for all Minisource microservices.

## Installation

```bash
go get github.com/minisource/go-common
```

## Package Index

| Package | Description |
|---------|-------------|
| `audit` | Audit logging utilities |
| `cache` | Redis caching helpers |
| `common` | Common utilities and helpers |
| `config` | Configuration loading |
| `constants` | Shared constants |
| `context` | Context utilities |
| `crypto` | Encryption and hashing |
| `db` | Database connection helpers |
| `dto` | Common data transfer objects |
| `errors` | Error handling utilities |
| `filter` | Query filtering helpers |
| `grpc` | gRPC server utilities |
| `grpcclient` | gRPC client helpers |
| `health` | Health check handlers |
| `http` | HTTP utilities and helpers |
| `httpclient` | HTTP client with retry/circuit breaker |
| `i18n` | Internationalization support |
| `limiter` | Rate limiting utilities |
| `logging` | Structured logging (zap) |
| `metrics` | Prometheus metrics |
| `pagination` | Pagination helpers |
| `repository` | Base repository patterns |
| `response` | API response builders |
| `service_errors` | Service error types |
| `shutdown` | Graceful shutdown |
| `testing` | Test utilities |
| `tracing` | OpenTelemetry tracing |
| `validations` | Input validation |

## Quick Start

### Logging

```go
import "github.com/minisource/go-common/logging"

// Initialize logger
logger := logging.NewLogger(&logging.LoggerConfig{
    FilePath: "logs/app.log",
    Level:    "info",
    Console:  true,
})

logger.Info(logging.General, logging.Api, "Server started", nil)
logger.Error(logging.Database, logging.Select, "Query failed", map[string]interface{}{
    "error": err.Error(),
})
```

### HTTP Response Builder

```go
import "github.com/minisource/go-common/response"

// Success response
response.NewBuilder(c).
    WithSuccess().
    WithData(user).
    WithMessage("User created").
    Send()

// Error response
response.NewBuilder(c).
    WithError(fiber.StatusBadRequest).
    WithMessage("Invalid input").
    WithValidationErrors(errors).
    Send()
```

### HTTP Client

```go
import "github.com/minisource/go-common/httpclient"

client := httpclient.NewClient(httpclient.Config{
    BaseURL:    "http://auth:9001",
    Timeout:    30 * time.Second,
    MaxRetries: 3,
})

resp, err := client.Get(ctx, "/api/v1/users/123", nil)
```

### gRPC Client

```go
import "github.com/minisource/go-common/grpcclient"

client, err := grpcclient.NewClient(ctx, grpcclient.Config{
    Address: "notifier:9003",
    TLS:     false,
})
defer client.Close()
```

### Middleware

```go
import (
    "github.com/minisource/go-common/http/middleware"
)

app := fiber.New()

// Request ID middleware
app.Use(middleware.RequestID())

// Auth middleware
app.Use(middleware.Auth(middleware.AuthConfig{
    TokenValidator: authClient.AsTokenValidator(),
}))

// Tenant middleware
app.Use(middleware.Tenant())

// Rate limiting
app.Use(middleware.RateLimiter(middleware.RateLimiterConfig{
    Max:      100,
    Duration: time.Minute,
}))

// Logging middleware
app.Use(middleware.Logger(logger))

// Recovery middleware
app.Use(middleware.Recovery())
```

### Error Handling

```go
import "github.com/minisource/go-common/service_errors"

// Create service error
err := service_errors.NewServiceError(
    service_errors.ErrInvalidInput,
    "email is required",
)

// Check error type
if service_errors.IsNotFound(err) {
    // Handle not found
}
```

### Pagination

```go
import "github.com/minisource/go-common/pagination"

// Parse pagination from request
page := pagination.FromFiberContext(c)

// Apply to query
query = query.Offset(page.Offset()).Limit(page.Limit())

// Build response
response := page.BuildResponse(items, totalCount)
```

### Validation

```go
import "github.com/minisource/go-common/validations"

type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

// Validate struct
errors := validations.Validate(req)
if len(errors) > 0 {
    // Handle validation errors
}
```

### Health Checks

```go
import "github.com/minisource/go-common/health"

app.Get("/health", health.Handler(health.Config{
    Checks: []health.Check{
        health.PostgresCheck(db),
        health.RedisCheck(redisClient),
    },
}))
```

### Metrics

```go
import "github.com/minisource/go-common/metrics"

// Initialize metrics
metrics.Init("my-service")

// Register with Fiber
app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
```

### Tracing

```go
import "github.com/minisource/go-common/tracing"

// Initialize tracer
tp, err := tracing.InitTracer(tracing.Config{
    ServiceName: "my-service",
    Endpoint:    "http://jaeger:14268/api/traces",
})
defer tp.Shutdown(ctx)

// Create span
ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()
```

### Graceful Shutdown

```go
import "github.com/minisource/go-common/shutdown"

shutdown.GracefulShutdown(ctx, shutdown.Config{
    Timeout: 30 * time.Second,
    Cleanup: func() error {
        db.Close()
        return nil
    },
})
```

## Middleware Reference

| Middleware | Description |
|------------|-------------|
| `Auth` | JWT authentication |
| `Tenant` | Multi-tenant context |
| `RequestID` | Request ID generation |
| `ContentType` | Content type validation |
| `Logger` | Request logging |
| `Recovery` | Panic recovery |
| `RateLimiter` | Rate limiting |
| `CORS` | CORS handling |
| `Security` | Security headers |
| `Tracing` | OpenTelemetry |
| `Prometheus` | Metrics collection |
| `Audit` | Audit logging |
| `Validation` | Request validation |
| `ServiceAuthRemote` | Service-to-service auth |

## Configuration

The `config` package supports loading from environment variables:

```go
import "github.com/minisource/go-common/config"

type AppConfig struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
}

cfg, err := config.Load[AppConfig]()
```

## Testing Utilities

```go
import "github.com/minisource/go-common/testing"

// Create test database
db := testing.NewTestDB(t)
defer db.Cleanup()

// Create test Redis
redis := testing.NewTestRedis(t)
defer redis.Cleanup()
```

## Contributing

1. Fork the repository
2. Create feature branch
3. Add tests for new functionality
4. Submit pull request

## License

MIT