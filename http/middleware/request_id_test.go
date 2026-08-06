package middleware

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRequestIDApp(t *testing.T, cfg RequestIDConfig) *fiber.App {
	app := fiber.New()
	app.Use(RequestID(cfg))
	app.Get("/ok", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"id": GetRequestID(c)})
	})
	return app
}

func TestRequestID_AcceptsValidIncoming(t *testing.T) {
	app := newRequestIDApp(t, DefaultRequestIDConfig())
	req := httptest.NewRequest("GET", "/ok", nil)
	req.Header.Set("X-Request-ID", "req-123-abc")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, "req-123-abc", resp.Header.Get("X-Request-ID"))

	body, _ := io.ReadAll(resp.Body)
	var out struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(body, &out))
	assert.Equal(t, "req-123-abc", out.ID)
}

func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	app := newRequestIDApp(t, DefaultRequestIDConfig())
	req := httptest.NewRequest("GET", "/ok", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	id := resp.Header.Get("X-Request-ID")
	assert.NotEmpty(t, id)

	body, _ := io.ReadAll(resp.Body)
	var out struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(body, &out))
	assert.Equal(t, id, out.ID)
}

func TestRequestID_RejectsOversized(t *testing.T) {
	cfg := DefaultRequestIDConfig()
	cfg.MaxLength = 16
	app := newRequestIDApp(t, cfg)

	req := httptest.NewRequest("GET", "/ok", nil)
	req.Header.Set("X-Request-ID", strings.Repeat("a", 64))
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	id := resp.Header.Get("X-Request-ID")
	assert.NotEmpty(t, id)
	assert.NotEqual(t, strings.Repeat("a", 64), id, "oversized input must be replaced with a generated ID")
	assert.Less(t, len(id), 64)
}

func TestIsValidRequestID_RejectsControlChars(t *testing.T) {
	// The HTTP layer rejects control characters before middleware runs, so the
	// predicate itself is unit-tested here.
	assert.False(t, isValidRequestID("bad\u0007id", 128))
	assert.False(t, isValidRequestID("with\u0001ctl", 128))
	assert.False(t, isValidRequestID("", 128))
	assert.False(t, isValidRequestID("too-long-value", 5))
	assert.True(t, isValidRequestID("req-123_abc.xyz", 128))
}

func TestRequestID_SkipsConfiguredPaths(t *testing.T) {
	cfg := DefaultRequestIDConfig()
	cfg.SkipPaths = []string{"/metrics"}
	app := fiber.New()
	app.Use(RequestID(cfg))
	app.Get("/metrics", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, "", resp.Header.Get("X-Request-ID"))
}
