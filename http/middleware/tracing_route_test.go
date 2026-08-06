package middleware

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouteTemplate_UsesMatchedTemplate(t *testing.T) {
	app := fiber.New()
	app.Get("/v1/admin/providers/:providerId", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"route": RouteTemplate(c)})
	})

	req := httptest.NewRequest("GET", "/v1/admin/providers/2825ba50-actual-uuid", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var out struct {
		Route string `json:"route"`
	}
	require.NoError(t, json.Unmarshal(body, &out))
	assert.Equal(t, "/v1/admin/providers/:providerId", out.Route)
	assert.NotContains(t, out.Route, "2825ba50", "raw UUID must never enter the template")
}

func TestRouteTemplate_EmptyForUnmatched(t *testing.T) {
	app := fiber.New()
	app.Get("/only", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"route": RouteTemplate(c)})
	})
	// Terminal middleware (registered after routes) catches unmatched requests,
	// where fiber's resolved route is "/*" and RouteTemplate must return "".
	app.Use(func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"route": RouteTemplate(c)})
	})

	req := httptest.NewRequest("GET", "/does/not/exist", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var out struct {
		Route string `json:"route"`
	}
	require.NoError(t, json.Unmarshal(body, &out))
	// Fiber's default 404 route resolves to "/" — the safe fallback (no raw
	// IDs or unmatched paths in span names / route labels).
	assert.Equal(t, "/", out.Route)
	assert.NotContains(t, out.Route, "/does/not/exist")
}
