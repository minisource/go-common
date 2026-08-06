package middleware

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureLogger records emitted records for assertions.
type captureLogger struct {
	records []logRecord
}

type logRecord struct {
	level string
	msg   string
	extra map[logging.ExtraKey]interface{}
}

func (l *captureLogger) Init() {}

func (l *captureLogger) Debug(cat logging.Category, sub logging.SubCategory, msg string, extra map[logging.ExtraKey]interface{}) {
	l.records = append(l.records, logRecord{"debug", msg, extra})
}
func (l *captureLogger) Info(cat logging.Category, sub logging.SubCategory, msg string, extra map[logging.ExtraKey]interface{}) {
	l.records = append(l.records, logRecord{"info", msg, extra})
}
func (l *captureLogger) Warn(cat logging.Category, sub logging.SubCategory, msg string, extra map[logging.ExtraKey]interface{}) {
	l.records = append(l.records, logRecord{"warn", msg, extra})
}
func (l *captureLogger) Error(cat logging.Category, sub logging.SubCategory, msg string, extra map[logging.ExtraKey]interface{}) {
	l.records = append(l.records, logRecord{"error", msg, extra})
}
func (l *captureLogger) Fatal(cat logging.Category, sub logging.SubCategory, msg string, extra map[logging.ExtraKey]interface{}) {
	l.records = append(l.records, logRecord{"fatal", msg, extra})
}
func (l *captureLogger) Debugf(template string, args ...interface{}) {}
func (l *captureLogger) Infof(template string, args ...interface{})  {}
func (l *captureLogger) Warnf(template string, args ...interface{})  {}
func (l *captureLogger) Errorf(template string, args ...interface{}) {}
func (l *captureLogger) Fatalf(template string, args ...interface{}) {}

func newAccessLogApp(cfg AccessLogConfig) (*fiber.App, *captureLogger) {
	logger := &captureLogger{}
	cfg.Logger = logger
	app := fiber.New()
	app.Use(RequestID())
	app.Use(AccessLog(cfg))
	handler := func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"id": c.Params("id")})
	}
	app.Get("/v1/users/:id", handler)
	app.Post("/v1/users/:id", handler)
	return app, logger
}

func TestAccessLog_CanonicalFields(t *testing.T) {
	app, logger := newAccessLogApp(DefaultAccessLogConfig("test-service"))

	req := httptest.NewRequest("GET", "/v1/users/abc-123", nil)
	req.Header.Set("X-Request-ID", "rid-1")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	require.Len(t, logger.records, 1)
	rec := logger.records[0]
	assert.Equal(t, AccessLogEvent, rec.msg)
	assert.Equal(t, "info", rec.level)
	assert.Equal(t, AccessLogEvent, rec.extra["event"])
	assert.Equal(t, "test-service", rec.extra["service"])
	assert.Equal(t, "GET", rec.extra["method"])
	assert.Equal(t, "/v1/users/:id", rec.extra["route"], "route must use the matched template")
	assert.Equal(t, "/v1/users/abc-123", rec.extra["path"])
	assert.Equal(t, 200, rec.extra["status_code"])
	assert.Equal(t, "rid-1", rec.extra["request_id"])
	assert.NotNil(t, rec.extra["duration_ms"])

	// No span exists (tracing middleware not mounted) → no fabricated IDs.
	assert.NotContains(t, rec.extra, "trace_id")
	assert.NotContains(t, rec.extra, "span_id")

	// Body capture off by default → no body fields.
	assert.NotContains(t, rec.extra, "request_body")
	assert.NotContains(t, rec.extra, "response_body")
}

func TestAccessLog_SkipsMetricsPath(t *testing.T) {
	cfg := DefaultAccessLogConfig("test-service")
	cfg.SkipPaths = []string{"/metrics"}
	app, logger := newAccessLogApp(cfg)
	app.Get("/metrics", func(c *fiber.Ctx) error { return c.SendString("ok") })

	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	assert.Len(t, logger.records, 0, "skip paths must not produce access records")
}

func TestAccessLog_RedactedRequestBodyWhenEnabled(t *testing.T) {
	cfg := DefaultAccessLogConfig("test-service")
	cfg.CaptureRequestBody = true
	app, logger := newAccessLogApp(cfg)

	req := httptest.NewRequest("POST", "/v1/users/abc-123",
		strings.NewReader(`{"email":"a@b.c","password":"supersecret","nested":{"otp":"123456"}}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	require.Len(t, logger.records, 1)
	body, ok := logger.records[0].extra["request_body"].(string)
	require.True(t, ok)
	assert.NotContains(t, body, "supersecret")
	assert.NotContains(t, body, "123456")
	assert.Contains(t, body, "[REDACTED]")
	assert.Contains(t, body, "a@b.c")
	assert.True(t, json.Valid([]byte(body)), "logged body must stay valid JSON")
}

func TestAccessLog_NonJSONBodyNotLoggedRaw(t *testing.T) {
	cfg := DefaultAccessLogConfig("test-service")
	cfg.CaptureRequestBody = true
	app, logger := newAccessLogApp(cfg)

	req := httptest.NewRequest("POST", "/v1/users/abc-123",
		strings.NewReader(`this is not json with secret=abc`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	require.Len(t, logger.records, 1)
	assert.NotContains(t, logger.records[0].extra, "request_body", "malformed JSON must never be logged raw")
}

func TestAccessLog_ResponseBodyCaptureWhenEnabled(t *testing.T) {
	cfg := DefaultAccessLogConfig("test-service")
	cfg.CaptureResponseBody = true
	app, logger := newAccessLogApp(cfg)

	req := httptest.NewRequest("GET", "/v1/users/abc-123", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	require.Len(t, logger.records, 1)
	body, ok := logger.records[0].extra["response_body"].(string)
	require.True(t, ok, "response body must be captured when enabled")
	assert.Contains(t, body, "abc-123")
	assert.True(t, json.Valid([]byte(body)), "logged response body must stay valid JSON")
	assert.NotEqual(t, "", body)
}

func TestAccessLog_DisabledByDefault(t *testing.T) {
	cfg := DefaultAccessLogConfig("test-service")
	app, logger := newAccessLogApp(cfg)

	req := httptest.NewRequest("POST", "/v1/users/abc-123",
		strings.NewReader(`{"password":"supersecret"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	require.Len(t, logger.records, 1)
	assert.NotContains(t, logger.records[0].extra, "request_body", "capture must be off by default")
	assert.NotContains(t, logger.records[0].extra, "response_body", "capture must be off by default")
}
