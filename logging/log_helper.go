package logging

import (
	"fmt"
	"runtime"
	"time"
)

func logParamsToZapParams(keys map[ExtraKey]interface{}) []interface{} {
	params := make([]interface{}, 0, len(keys))

	for k, v := range keys {
		params = append(params, string(k))
		params = append(params, v)
	}

	return params
}

func logParamsToZeroParams(keys map[ExtraKey]interface{}) map[string]interface{} {
	params := map[string]interface{}{}

	for k, v := range keys {
		params[string(k)] = v
	}

	return params
}

// prepareLogInfo prepares log information with category and extra fields
func prepareLogInfo(cat Category, sub SubCategory, extra map[ExtraKey]interface{}) []interface{} {
	if extra == nil {
		extra = make(map[ExtraKey]interface{})
	}

	extra["Category"] = cat
	extra["SubCategory"] = sub

	return logParamsToZapParams(extra)
}

// GetCallerInfo returns caller function information for debug logging
func GetCallerInfo(skip int) string {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return fmt.Sprintf("%s:%d", file, line)
	}
	return fmt.Sprintf("%s:%d %s", file, line, fn.Name())
}

// DebugContext adds contextual debug information
type DebugContext struct {
	RequestID string
	UserID    string
	Method    string
	Path      string
	StartTime time.Time
	Extra     map[string]interface{}
}

// ToMap converts debug context to map for logging
func (d *DebugContext) ToMap() map[ExtraKey]interface{} {
	m := make(map[ExtraKey]interface{})

	if d.RequestID != "" {
		m["request_id"] = d.RequestID
	}
	if d.UserID != "" {
		m["user_id"] = d.UserID
	}
	if d.Method != "" {
		m["method"] = d.Method
	}
	if d.Path != "" {
		m["path"] = d.Path
	}
	if !d.StartTime.IsZero() {
		m["elapsed"] = time.Since(d.StartTime).Milliseconds()
	}

	for k, v := range d.Extra {
		m[ExtraKey(k)] = v
	}

	return m
}

// WithExtra adds extra fields to debug context
func (d *DebugContext) WithExtra(key string, value interface{}) *DebugContext {
	if d.Extra == nil {
		d.Extra = make(map[string]interface{})
	}
	d.Extra[key] = value
	return d
}
