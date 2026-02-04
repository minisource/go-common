# Centralized HTTP/gRPC Client System

This package provides a centralized, reusable client system for microservice communication following SOLID principles.

## Features

- **Automatic Retry Logic**: Exponential backoff with configurable max retries
- **Debug Logging**: Comprehensive logging at every stage (request, retry, response, errors)
- **Error Handling**: Proper error types and context
- **i18n Support**: Through integrated logging system
- **Circuit Breaker Ready**: Designed to integrate with circuit breaker pattern
- **Protocol Agnostic**: Supports both HTTP and gRPC
- **Thread-Safe**: Safe for concurrent use
- **SOLID Principles**: Clean architecture with separation of concerns

## Architecture

```
go-common/
├── httpclient/          # HTTP client implementation
│   ├── client.go        # Main HTTP client with retry and logging
│   └── errors.go        # HTTP-specific errors
├── grpcclient/          # gRPC client implementation
│   ├── client.go        # Main gRPC client with interceptors
│   └── errors.go        # gRPC-specific errors
└── logging/             # Logging framework (existing)
```

## HTTP Client Usage

### Basic Example

```go
import (
    "context"
    "github.com/minisource/go-common/httpclient"
    "github.com/minisource/go-common/logging"
)

// Create logger
logger := logging.NewLogger(logging.LogConfig{
    Level:      logging.DebugLevel,  // Use DebugLevel to see retry logs
    OutputPath: "stdout",
})

// Create HTTP client
client := httpclient.NewClient(httpclient.Config{
    BaseURL:     "http://api.example.com",
    ServiceName: "example-service",
    Timeout:     30 * time.Second,
    RetryConfig: httpclient.DefaultRetryConfig(), // 3 retries with exponential backoff
    Logger:      logger,
})

// Make a request
resp, err := client.Post(ctx, "/api/users", map[string]string{
    "name": "John",
}, nil)

if err != nil {
    // Check if service is unavailable
    if errors.As(err, &httpclient.ServiceUnavailableError{}) {
        log.Println("Service is down")
    }
    return err
}

// Decode response
var user User
if err := resp.DecodeJSON(&user); err != nil {
    return err
}
```

### Custom Retry Configuration

```go
retryConfig := httpclient.RetryConfig{
    MaxRetries:    5,
    InitialDelay:  1 * time.Second,
    MaxDelay:      10 * time.Second,
    BackoffFactor: 2.0,
    RetryableErrors: []int{
        http.StatusServiceUnavailable,
        http.StatusGatewayTimeout,
    },
}

client := httpclient.NewClient(httpclient.Config{
    BaseURL:     "http://api.example.com",
    ServiceName: "example-service",
    RetryConfig: retryConfig,
    Logger:      logger,
})
```

### Using Interceptors

```go
// Add auth header to all requests
authInterceptor := func(ctx context.Context, req *http.Request) error {
    token := getTokenFromContext(ctx)
    req.Header.Set("Authorization", "Bearer " + token)
    return nil
}

client := httpclient.NewClient(httpclient.Config{
    BaseURL:      "http://api.example.com",
    ServiceName:  "example-service",
    Interceptors: []httpclient.Interceptor{authInterceptor},
    Logger:       logger,
})
```

## gRPC Client Usage

### Basic Example

```go
import (
    "context"
    "github.com/minisource/go-common/grpcclient"
    "github.com/minisource/go-common/logging"
)

// Create logger
logger := logging.NewLogger(logging.LogConfig{
    Level:      logging.DebugLevel,
    OutputPath: "stdout",
})

// Create gRPC client
grpcClient, err := grpcclient.NewClient(ctx, grpcclient.Config{
    Target:      "localhost:9002",
    ServiceName: "notifier-service",
    RetryConfig: grpcclient.DefaultRetryConfig(),
    Logger:      logger,
})
if err != nil {
    return err
}
defer grpcClient.Close()

// Use the connection with your generated gRPC clients
conn := grpcClient.Conn()
myService := pb.NewMyServiceClient(conn)
```

### With Bearer Authentication

```go
// Get token from auth service
token := "your-jwt-token"

grpcClient, err := grpcclient.NewClient(ctx, grpcclient.Config{
    Target:      "localhost:9002",
    ServiceName: "notifier-service",
    Interceptors: []grpc.UnaryClientInterceptor{
        grpcclient.BearerAuthInterceptor(token),
    },
    StreamInterceptors: []grpc.StreamClientInterceptor{
        grpcclient.BearerAuthStreamInterceptor(token),
    },
    Logger: logger,
})
```

### Dynamic Token Authentication

```go
// For dynamically refreshing tokens
authInterceptor := func(ctx context.Context, method string, req, reply interface{}, 
    cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
    
    // Get fresh token
    token, err := authClient.GetToken(ctx)
    if err != nil {
        return err
    }
    
    // Use built-in bearer auth interceptor
    return grpcclient.BearerAuthInterceptor(token)(ctx, method, req, reply, cc, invoker, opts...)
}

grpcClient, err := grpcclient.NewClient(ctx, grpcclient.Config{
    Target:       "localhost:9002",
    ServiceName:  "notifier-service",
    Interceptors: []grpc.UnaryClientInterceptor{authInterceptor},
    Logger:       logger,
})
```

## Debug Logging

The client system provides comprehensive debug logging at every stage:

### HTTP Client Logs

```
[DEBUG] Starting HTTP request (service: auth-service, method: POST, path: /api/v1/service/auth)
[DEBUG] Request body (service: auth-service, body: {"clientId":"...","clientSecret":"..."})
[DEBUG] Response received (service: auth-service, statusCode: 200)
[INFO]  HTTP request completed (service: auth-service, duration: 123ms, attempt: 1)
```

### Retry Logs

```
[WARN]  HTTP request returned retryable error (service: notifier-service, statusCode: 503, attempt: 1)
[DEBUG] Retrying request (service: notifier-service, attempt: 2, delay: 500ms)
[WARN]  HTTP request failed (service: notifier-service, attempt: 2, error: connection refused)
[DEBUG] Retrying request (service: notifier-service, attempt: 3, delay: 1s)
[INFO]  HTTP request completed (service: notifier-service, statusCode: 200, attempt: 3)
```

### gRPC Client Logs

```
[INFO]  gRPC connection established (service: notifier-service, target: localhost:9002)
[DEBUG] Starting gRPC request (service: notifier-service, method: /notifier.v1.NotificationService/SendSMS)
[DEBUG] Adding auth token to request (method: /notifier.v1.NotificationService/SendSMS)
[INFO]  gRPC request completed (service: notifier-service, method: /notifier.v1.NotificationService/SendSMS, duration: 45ms)
```

## Error Handling

### HTTP Errors

```go
resp, err := client.Post(ctx, "/api/users", userData, nil)
if err != nil {
    // Check for service unavailable (after all retries exhausted)
    var svcErr *httpclient.ServiceUnavailableError
    if errors.As(err, &svcErr) {
        // Service is down, show friendly error to user
        return fmt.Errorf("The %s is temporarily unavailable", svcErr.ServiceName)
    }
    return err
}
```

### gRPC Errors

```go
resp, err := client.SendSMS(ctx, req)
if err != nil {
    // Check for service unavailable
    var svcErr *grpcclient.ServiceUnavailableError
    if errors.As(err, &svcErr) {
        return fmt.Errorf("The %s is temporarily unavailable", svcErr.ServiceName)
    }
    return err
}
```

## SOLID Principles Implementation

### Single Responsibility Principle (SRP)
- **HTTP Client**: Only handles HTTP communication
- **gRPC Client**: Only handles gRPC communication
- **Retry Logic**: Separated into its own interceptor/middleware
- **Logging**: Separated into logging interceptors
- **Auth**: Separated into auth interceptors

### Open/Closed Principle (OCP)
- Clients are open for extension through **interceptors**
- Closed for modification - add new behaviors without changing core code
- Example: Add metrics, tracing, or custom headers via interceptors

### Liskov Substitution Principle (LSP)
- Both HTTP and gRPC clients implement similar patterns
- Can be swapped without breaking dependent code
- Consistent error handling across protocols

### Interface Segregation Principle (ISP)
- Small, focused interfaces: `Interceptor`, `Logger`
- Clients don't depend on interfaces they don't use
- Each component has minimal, specific responsibilities

### Dependency Inversion Principle (DIP)
- Clients depend on abstractions (Logger interface), not concrete types
- Interceptors can be injected, promoting testability
- Configuration-driven behavior

## Integration with go-sdk

### Auth SDK Client

The auth SDK client now uses the common HTTP client:

```go
import (
    "github.com/minisource/go-sdk/auth"
    "github.com/minisource/go-common/logging"
)

logger := logging.NewLogger(logging.LogConfig{
    Level:      logging.DebugLevel,
    OutputPath: "stdout",
})

client := auth.NewClient(auth.ClientConfig{
    BaseURL:      "http://localhost:9001",
    ClientID:     "my-service",
    ClientSecret: "secret",
    Logger:       logger,  // Debug logs enabled
})
```

### Notifier SDK Client

The notifier SDK client now uses the common gRPC client:

```go
import (
    "github.com/minisource/go-sdk/notifier"
    "github.com/minisource/go-common/logging"
)

logger := logging.NewLogger(logging.LogConfig{
    Level:      logging.DebugLevel,
    OutputPath: "stdout",
})

client, err := notifier.NewClient(ctx, notifier.Config{
    Address:    "localhost:9002",
    AuthClient: authClient,
    Logger:     logger,  // Debug logs enabled
})
```

## Future Enhancements

1. **Circuit Breaker Integration**: Add circuit breaker middleware
2. **Metrics Collection**: Add Prometheus/OpenTelemetry support
3. **Distributed Tracing**: Add trace ID propagation
4. **Rate Limiting**: Add client-side rate limiting
5. **Request/Response Caching**: Add caching layer for idempotent requests

## Benefits

✅ **No Code Duplication**: Single implementation for all microservices
✅ **Consistent Logging**: Same debug logs across all services
✅ **Easy Debugging**: Trace every request and retry
✅ **Testable**: Easy to mock and test
✅ **Maintainable**: Changes in one place affect all services
✅ **Production-Ready**: Retry logic, error handling, logging built-in

## Migration Guide

### Before (Custom Implementation)

```go
// Custom HTTP client with manual retry logic
for attempt := 0; attempt < 3; attempt++ {
    resp, err := http.Post(url, body)
    if err == nil {
        break
    }
    time.Sleep(time.Second * attempt)
}
```

### After (Common Client)

```go
// Automatic retry with exponential backoff and logging
client := httpclient.NewClient(config)
resp, err := client.Post(ctx, "/api/users", body, nil)
```

**Result**: 50+ lines of retry logic → 2 lines with better error handling and logging!
