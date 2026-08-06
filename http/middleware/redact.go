package middleware

import (
	"encoding/json"
	"strings"
)

// RedactedValue is the placeholder substituted for sensitive values.
const RedactedValue = "[REDACTED]"

// DefaultRedactFields returns the default list of sensitive JSON keys that are
// redacted before any body is logged. Matching is case-insensitive.
func DefaultRedactFields() []string {
	return []string{
		"password",
		"password_confirmation",
		"current_password",
		"new_password",
		"otp",
		"code",
		"token",
		"access_token",
		"refresh_token",
		"id_token",
		"authorization",
		"cookie",
		"set-cookie",
		"api_key",
		"apikey",
		"secret",
		"client_secret",
		"provider_secret",
	}
}

// isSensitiveKey reports whether key matches a sensitive field name,
// case-insensitively.
func isSensitiveKey(key string, sensitive []string) bool {
	lk := strings.ToLower(key)
	for _, s := range sensitive {
		if lk == strings.ToLower(s) {
			return true
		}
	}
	return false
}

// RedactJSON redacts values under sensitive keys (recursively, including
// nested objects and arrays) and returns the re-encoded JSON. It returns nil
// when the input is not valid JSON so callers never log an unsafe raw body.
func RedactJSON(body []byte, sensitive []string) []byte {
	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		return nil
	}
	redactValue(v, sensitive)
	out, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return out
}

// redactValue walks a decoded JSON tree and replaces values whose keys match a
// sensitive field. A redacted value is replaced wholesale (never recursed into)
// so nested secrets cannot leak.
func redactValue(v interface{}, sensitive []string) {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, val := range t {
			if isSensitiveKey(k, sensitive) {
				t[k] = RedactedValue
			} else {
				redactValue(val, sensitive)
			}
		}
	case []interface{}:
		for _, item := range t {
			redactValue(item, sensitive)
		}
	}
}
