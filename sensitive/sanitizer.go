// Package sensitive provides utilities to redact sensitive data (passwords, OTP codes,
// tokens, secrets, etc.) from logs and database storage when the application is NOT
// running in development mode.
//
// Usage:
//
//	import "github.com/minisource/go-common/sensitive"
//
//	// Redact a value in a log extra map:
//	logger.Info(cat, sub, "msg", sensitive.LogMap(extra, "code", "password"))
package sensitive

import (
	"os"
	"strings"

	"github.com/minisource/go-common/logging"
)

// RedactedPlaceholder is the string that replaces sensitive values in non-dev environments.
const RedactedPlaceholder = "***"

// Common sensitive ExtraKey constants for convenience.
var (
	KeyCode      = logging.ExtraKey("code")
	KeyPassword  = logging.ExtraKey("password")
	KeyToken     = logging.ExtraKey("token")
	KeySecret    = logging.ExtraKey("secret")
	KeyAccessToken  = logging.ExtraKey("accessToken")
	KeyRefreshToken = logging.ExtraKey("refreshToken")
)

// Value returns the interface unchanged in development mode, or RedactedPlaceholder otherwise.
// Use this when logging individual sensitive values outside a map context.
func Value(v interface{}) interface{} {
	if isDev() {
		return v
	}
	return RedactedPlaceholder
}

// String returns the string unchanged in development mode, or RedactedPlaceholder otherwise.
func String(s string) string {
	if isDev() {
		return s
	}
	return RedactedPlaceholder
}

// LogMap returns a copy of extra with the given sensitive keys replaced by RedactedPlaceholder
// when NOT in development mode. The original map is not modified.
//
// Keys are matched by their string representation, so both ExtraKey and plain string work:
//
//	extra := map[logging.ExtraKey]interface{}{"code": "123456", "target": "user@x.com"}
//	safe := sensitive.LogMap(extra, "code", "otpCode")
//	logger.Info(cat, sub, "msg", safe)
func LogMap(extra map[logging.ExtraKey]interface{}, sensitiveKeys ...logging.ExtraKey) map[logging.ExtraKey]interface{} {
	if extra == nil || len(sensitiveKeys) == 0 {
		return extra
	}
	if isDev() {
		return extra
	}

	keySet := make(map[logging.ExtraKey]bool, len(sensitiveKeys))
	for _, k := range sensitiveKeys {
		keySet[k] = true
	}

	result := make(map[logging.ExtraKey]interface{}, len(extra))
	for k, v := range extra {
		if keySet[k] {
			result[k] = RedactedPlaceholder
		} else {
			result[k] = v
		}
	}
	return result
}

// isDev returns true if the app is running in development or test mode.
func isDev() bool {
	env := strings.ToLower(os.Getenv("SERVER_MODE"))
	if env == "" {
		env = strings.ToLower(os.Getenv("APP_ENV"))
	}
	if env == "" {
		env = strings.ToLower(os.Getenv("ENV"))
	}
	return env == "development" || env == "dev" || env == "test"
}
