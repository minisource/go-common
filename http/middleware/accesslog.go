package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/logging"
	"go.opentelemetry.io/otel/trace"
)

// AccessLogEvent is the canonical request-completion event name shared by all
// MiniSource services.
const AccessLogEvent = "http_request_completed"

// AccessLogConfig configures the shared structured HTTP access log middleware.
//
// Body capture is OFF by default and must be enabled explicitly. When enabled
// it is bounded, content-type gated, and redacted; it is never emitted as span
// attributes.
type AccessLogConfig struct {
	Service     string
	Environment string

	// Logger is the go-common logger. When nil the middleware writes the same
	// canonical JSON to stdout (used by services with their own logger stack).
	Logger logging.Logger

	// SkipPaths are exact paths excluded from access logging (noise/health).
	SkipPaths []string

	// Enabled disables the whole middleware (default true).
	Enabled bool

	// Body capture options (all default OFF).
	CaptureRequestBody  bool
	CaptureResponseBody bool
	MaxBodySizeBytes    int
	AllowedContentTypes []string
	RedactFields        []string
}

// DefaultAccessLogConfig returns safe defaults: metadata only, no body capture.
func DefaultAccessLogConfig(service string) AccessLogConfig {
	return AccessLogConfig{
		Service:             service,
		Environment:         "development",
		SkipPaths:           []string{"/metrics", "/swagger"},
		Enabled:             true,
		MaxBodySizeBytes:    8192,
		AllowedContentTypes: []string{"application/json"},
		RedactFields:        DefaultRedactFields(),
	}
}

// LoadAccessLogConfigFromEnv builds a config from OBSERVABILITY_* environment
// variables. Body capture remains disabled unless explicitly enabled.
func LoadAccessLogConfigFromEnv(service string, logger logging.Logger) AccessLogConfig {
	cfg := DefaultAccessLogConfig(service)
	cfg.Logger = logger

	if v := os.Getenv("ENVIRONMENT"); v != "" {
		cfg.Environment = v
	}
	if v := os.Getenv("APP_ENV"); v != "" {
		cfg.Environment = v
	}
	if v := os.Getenv("OBSERVABILITY_HTTP_LOGGING_ENABLED"); v != "" {
		cfg.Enabled = parseBool(v, cfg.Enabled)
	}
	if v := os.Getenv("OBSERVABILITY_CAPTURE_REQUEST_BODY"); v != "" {
		cfg.CaptureRequestBody = parseBool(v, false)
	}
	if v := os.Getenv("OBSERVABILITY_CAPTURE_RESPONSE_BODY"); v != "" {
		cfg.CaptureResponseBody = parseBool(v, false)
	}
	if v := os.Getenv("OBSERVABILITY_MAX_BODY_SIZE_BYTES"); v != "" {
		if n := parseInt(v); n > 0 {
			cfg.MaxBodySizeBytes = n
		}
	}
	if v := os.Getenv("OBSERVABILITY_ALLOWED_CONTENT_TYPES"); v != "" {
		cfg.AllowedContentTypes = strings.Split(v, ",")
	}
	return cfg
}

// AccessLog returns the shared structured access log middleware. It emits one
// canonical event per request with trace_id/span_id/request_id/method/route/
// status/duration/sizes and (only when enabled) redacted bounded bodies.
func AccessLog(cfg AccessLogConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !cfg.Enabled {
			return c.Next()
		}
		for _, p := range cfg.SkipPaths {
			if c.Path() == p {
				return c.Next()
			}
		}

		start := time.Now()

		// Optional, bounded, sanitized request body capture. Reading the body
		// does not consume it (fasthttp keeps it buffered), so handlers can
		// still read the request body normally.
		var requestBody string
		if cfg.CaptureRequestBody && contentTypeAllowed(c.Request().Header.ContentType(), cfg.AllowedContentTypes) {
			requestBody = captureBody(c.Request().Body(), cfg.MaxBodySizeBytes, cfg.RedactFields)
		}

		err := c.Next()

		// Optional, bounded, sanitized response body capture. Read directly
		// from the completed response — never via a pre-set body stream
		// (handlers replace the stream, which would make capture dead code).
		// Not applied to WebSocket upgrades.
		var responseBody string
		if cfg.CaptureResponseBody && !isUpgradeRequest(c) && contentTypeAllowed(c.Response().Header.ContentType(), cfg.AllowedContentTypes) {
			responseBody = captureBody(c.Response().Body(), cfg.MaxBodySizeBytes, cfg.RedactFields)
		}

		statusCode := c.Response().StatusCode()
		durationMs := float64(time.Since(start).Nanoseconds()) / 1e6

		fields := buildAccessLogFields(cfg, c, statusCode, durationMs, requestBody, responseBody, err)
		level := "info"
		if statusCode >= 500 {
			level = "error"
		} else if statusCode >= 400 {
			level = "warn"
		}

		if cfg.Logger != nil {
			emitViaLogger(cfg.Logger, level, fields)
		} else {
			emitStdoutJSON(level, fields)
		}

		return err
	}
}

// buildAccessLogFields composes the canonical field set.
func buildAccessLogFields(
	cfg AccessLogConfig, c *fiber.Ctx, statusCode int, durationMs float64,
	requestBody, responseBody string, err error,
) map[logging.ExtraKey]interface{} {
	fields := map[logging.ExtraKey]interface{}{
		"event":            AccessLogEvent,
		"service":          cfg.Service,
		"method":           c.Method(),
		"path":             c.Path(),
		"status_code":      statusCode,
		"duration_ms":      durationMs,
		"client_address":   c.IP(),
		"user_agent":       string(c.Request().Header.UserAgent()),
	}
	if cfg.Environment != "" {
		fields["environment"] = cfg.Environment
	}
	if rid := GetRequestID(c); rid != "" {
		fields["request_id"] = rid
	}

	// Trace/span IDs from the active span (empty when tracing is disabled;
	// never fabricated).
	if sc := trace.SpanFromContext(c.UserContext()).SpanContext(); sc.HasTraceID() {
		fields["trace_id"] = sc.TraceID().String()
		fields["span_id"] = sc.SpanID().String()
	}

	// Matched route template (e.g. /v1/users/:id), falling back to the path.
	if route := RouteTemplate(c); route != "" {
		fields["route"] = route
	} else {
		fields["route"] = c.Path()
	}

	if n := len(c.Request().Body()); n > 0 {
		fields["request_size_bytes"] = n
	}
	if n := len(c.Response().Body()); n > 0 {
		fields["response_size_bytes"] = n
	}

	if err != nil {
		fields["error_type"] = "handler_error"
	} else if statusCode >= 500 {
		fields["error_type"] = httpStatusText(statusCode)
	}

	if requestBody != "" {
		fields["request_body"] = requestBody
	}
	if responseBody != "" {
		fields["response_body"] = responseBody
	}

	return fields
}

func emitViaLogger(logger logging.Logger, level string, fields map[logging.ExtraKey]interface{}) {
	switch level {
	case "error":
		logger.Error(logging.General, logging.Api, AccessLogEvent, fields)
	case "warn":
		logger.Warn(logging.General, logging.Api, AccessLogEvent, fields)
	default:
		logger.Info(logging.General, logging.Api, AccessLogEvent, fields)
	}
}

// emitStdoutJSON writes the canonical event as JSON to stdout (gateway's
// logger stack is not the go-common logger). Shape mirrors zap output so
// Promtail/Loki parsing stays uniform.
func emitStdoutJSON(level string, fields map[logging.ExtraKey]interface{}) {
	rec := map[string]interface{}{
		"level": level,
		"ts":    time.Now().Format(time.RFC3339Nano),
		"msg":   AccessLogEvent,
	}
	for k, v := range fields {
		rec[string(k)] = v
	}
	b, _ := json.Marshal(rec)
	fmt.Fprintln(os.Stdout, string(b))
}

// --- body capture helpers -------------------------------------------------

// contentTypeAllowed reports whether a media type is in the allowed set
// (checked against the media type only; multipart and binary are never
// allowed).
func contentTypeAllowed(rawCT []byte, allowed []string) bool {
	ct := strings.ToLower(strings.TrimSpace(string(rawCT)))
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = strings.TrimSpace(ct[:i])
	}
	if ct == "" || strings.HasPrefix(ct, "multipart/") {
		return false
	}
	for _, a := range allowed {
		if ct == strings.ToLower(strings.TrimSpace(a)) {
			return true
		}
	}
	return false
}

func isUpgradeRequest(c *fiber.Ctx) bool {
	return strings.EqualFold(strings.TrimSpace(c.Get("Upgrade")), "websocket")
}

// captureBody copies at most max bytes of body and redacts sensitive fields.
// Malformed/non-JSON bodies yield "" (never logged raw).
func captureBody(body []byte, max int, redact []string) string {
	if len(body) == 0 {
		return ""
	}
	truncated := len(body) > max
	if truncated {
		body = body[:max]
	}
	out := RedactJSON(body, redact)
	if out == nil {
		return ""
	}
	s := string(out)
	if truncated {
		s += " …truncated"
	}
	return s
}

// cappedBodyReader is a legacy io.Reader/Writer kept for ResponseWriter-style
// compatibility; current response capture reads the completed response body
// directly, so this type is only retained for API stability.
type cappedBodyReader struct {
	body  bytes.Buffer
	limit int
	full  bool
}

func newCappedBodyReader(limit int) *cappedBodyReader {
	return &cappedBodyReader{limit: limit}
}

// Read implements io.Reader.
func (b *cappedBodyReader) Read(p []byte) (int, error) {
	return b.body.Read(p)
}

// Write implements io.Writer; caps the internal buffer and reports the full
// write length so the response is never truncated for the client.
func (b *cappedBodyReader) Write(p []byte) (int, error) {
	if room := b.limit - b.body.Len(); room > 0 {
		if len(p) > room {
			b.body.Write(p[:room])
			b.full = true
		} else {
			b.body.Write(p)
		}
	} else if len(p) > 0 {
		b.full = true
	}
	return len(p), nil
}

// Bytes returns the captured (possibly truncated) content.
func (b *cappedBodyReader) Bytes() []byte {
	return b.body.Bytes()
}

// Truncated reports whether more data was written than the capture limit.
func (b *cappedBodyReader) Truncated() bool { return b.full }

var _ io.Reader = (*cappedBodyReader)(nil)
var _ io.Writer = (*cappedBodyReader)(nil)

func httpStatusText(code int) string {
	if t := http.StatusText(code); t != "" {
		return t
	}
	return fmt.Sprintf("status_%d", code)
}

func parseBool(v string, def bool) bool {
	b, err := strconv.ParseBool(strings.TrimSpace(v))
	if err != nil {
		return def
	}
	return b
}

func parseInt(v string) int {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return 0
	}
	return n
}
