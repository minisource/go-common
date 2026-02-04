# I18n and Enhanced Error Handling

This document describes the new i18n (internationalization) support and enhanced error handling features in the common_go library.

## Features

### 1. Internationalization (i18n)

Support for multiple languages with embedded translation files.

**Supported Languages:**
- English (en)
- Persian/Farsi (fa)

**Translation Categories:**
- `errors.*` - Error messages
- `notifications.*` - Notification-related messages
- `validation.*` - Validation messages
- `common.*` - Common UI messages

### 2. Environment-Aware Error Details

Error responses automatically adjust based on the environment:

- **Development Mode**: Returns detailed error information including technical messages, stack traces, and debug data
- **Production Mode**: Returns only user-friendly messages

### 3. Enhanced Debug Logging

Improved logging with contextual information including:
- Request ID
- User ID
- HTTP method and path
- Execution time
- Caller information
- Custom debug fields

## Usage

### Basic I18n Usage

```go
import "github.com/minisource/go-common/i18n"

// Simple translation
message := i18n.T(ctx, "errors.not_found")

// Translation with parameters
message := i18n.T(ctx, "validation.min_length", map[string]interface{}{
    "Min": 5,
})

// With specific language
message := i18n.TLang("fa", "errors.unauthorized")
```

### Language Detection

The i18n system detects language in this priority order:

1. **Query Parameter**: `?lang=fa` or `?lang=en`
2. **Accept-Language Header**: `Accept-Language: fa, en;q=0.9`
3. **Default Language**: `en`

Example requests:
```bash
# Use Persian
curl -H "Accept-Language: fa" http://localhost:8080/api/v1/notifications

# Or via query parameter
curl http://localhost:8080/api/v1/notifications?lang=fa
```

### Error Handling with I18n

```go
import (
    "github.com/minisource/go-common/common"
    "github.com/minisource/go-common/http/helper"
    "github.com/minisource/go-common/service_errors"
)

// Create a service error with code
err := service_errors.NewServiceError(
    "notification_not_found",
    "Notification not found",
    "Database query returned no results",
).WithDetails(map[string]interface{}{
    "notification_id": id,
})

// Return error response (automatically translated)
return c.Status(404).JSON(
    helper.GenerateBaseResponseWithServiceError(
        c,                      // Fiber context for language detection
        nil,                    // result
        false,                  // success
        helper.NotFoundError,   // result code
        err,                    // service error
        common.IsDevelopment(), // show detailed errors in dev mode
    ),
)
```

### Environment Detection

```go
import "github.com/minisource/go-common/common"

// Check environment
if common.IsDevelopment() {
    // Add debug logging
    logger.Debug(logging.Api, logging.Internal, "Debug info", extraData)
}

if common.IsProduction() {
    // Production-only code
}

// Get environment
env := common.GetEnvironment() // "development", "staging", "production", "test"

// Should show detailed errors?
showDetails := common.ShouldShowDetailedErrors()
```

**Environment Variables:**
Set one of these environment variables:
- `APP_ENV=development`
- `ENV=production`
- `ENVIRONMENT=staging`

Values: `development`, `dev`, `staging`, `stage`, `production`, `prod`, `test`, `testing`

### Debug Logging

```go
import "github.com/minisource/go-common/logging"

// Create debug context
debugCtx := &logging.DebugContext{
    RequestID: c.Get("X-Request-ID"),
    UserID:    c.Get("X-User-ID"),
    Method:    c.Method(),
    Path:      c.Path(),
}

// Add extra fields
debugCtx.WithExtra("notification_id", id).
         WithExtra("retry_count", retries)

// Log with context
logger.Debug(
    logging.Api,
    logging.Internal,
    "Processing notification",
    debugCtx.ToMap(),
)
```

### Response Helpers

```go
import "github.com/minisource/go-common/http/helper"

// 1. Simple i18n response
helper.GenerateI18nResponse(
    c,                               // context
    result,                          // data
    true,                            // success
    0,                               // result code
    "notifications.notification_created", // i18n key
)

// 2. I18n response with parameters
helper.GenerateI18nResponse(
    c,
    result,
    false,
    helper.ValidationError,
    "validation.min_length",
    map[string]interface{}{"Min": 5},
)

// 3. Service error response (environment-aware)
helper.GenerateBaseResponseWithServiceError(
    c,
    nil,
    false,
    helper.NotFoundError,
    serviceError,
    common.IsDevelopment(),
)
```

## Example Responses

### Development Mode Response

```json
{
  "result": null,
  "success": false,
  "resultCode": 40401,
  "message": "Notification not found",
  "error": {
    "message": "Notification not found",
    "code": "notification_not_found",
    "technical_message": "Database query returned no results",
    "error": "record not found",
    "details": {
      "notification_id": "123e4567-e89b-12d3-a456-426614174000"
    },
    "stack": "goroutine 1 [running]..."
  }
}
```

### Production Mode Response

```json
{
  "result": null,
  "success": false,
  "resultCode": 40401,
  "message": "Notification not found"
}
```

### Persian Language Response

```json
{
  "result": null,
  "success": false,
  "resultCode": 40401,
  "message": "اعلان یافت نشد"
}
```

## Adding New Translations

### 1. Add to English (en.json)

```json
{
  "errors": {
    "new_error": "This is a new error message"
  }
}
```

### 2. Add to Persian (fa.json)

```json
{
  "errors": {
    "new_error": "این یک پیام خطای جدید است"
  }
}
```

### 3. Use in Code

```go
message := i18n.T(ctx, "errors.new_error")
```

## Error Code Conventions

Use snake_case for error codes to match i18n keys:

```go
const (
    NotificationNotFound = "notification_not_found"  // ✅ Good
    NotFound = "not_found"                           // ✅ Good
    
    NotificationNotFound = "NotificationNotFound"    // ❌ Bad
    NotFound = "Not Found"                           // ❌ Bad
)
```

## Best Practices

1. **Always use i18n keys** instead of hardcoded messages
2. **Use environment detection** to control error detail levels
3. **Add debug logging** in development mode for troubleshooting
4. **Use snake_case** for error codes to match i18n keys
5. **Keep translations consistent** across languages
6. **Test with different languages** using `?lang=fa` parameter

## Complete Handler Example

See [example_i18n_handler.go](../../../notifier/api/v1/handlers/example_i18n_handler.go) for a complete implementation example.

## Migration Guide

### Old Code

```go
return c.Status(404).JSON(
    helper.GenerateBaseResponseWithError(
        nil,
        false,
        helper.NotFoundError,
        errors.New("Notification not found"),
    ),
)
```

### New Code

```go
err := service.NewNotificationNotFoundError().
    WithDetails(map[string]interface{}{"id": id})

return c.Status(404).JSON(
    helper.GenerateBaseResponseWithServiceError(
        c,
        nil,
        false,
        helper.NotFoundError,
        err,
        common.IsDevelopment(),
    ),
)
```

## Environment Configuration

Create a `.env` file:

```env
# Development
APP_ENV=development

# Staging
APP_ENV=staging

# Production
APP_ENV=production
```

## Supported Languages

| Code | Language | Status |
|------|----------|--------|
| en   | English  | ✅ Complete |
| fa   | Persian/Farsi | ✅ Complete |

To add more languages, create a new JSON file in `common_go/i18n/locales/` and update the translator.
