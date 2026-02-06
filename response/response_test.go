package response

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestApp() *fiber.App {
	return fiber.New()
}

func TestNewBuilder(t *testing.T) {
	app := setupTestApp()

	app.Get("/test", func(c *fiber.Ctx) error {
		return New().
			Success(true).
			Data(map[string]string{"message": "Test successful"}).
			Send(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result Response
	json.Unmarshal(body, &result)

	assert.True(t, result.Success)
	assert.NotNil(t, result.Data)
}

func TestBuilderWithData(t *testing.T) {
	app := setupTestApp()

	testData := map[string]string{"key": "value"}

	app.Get("/test", func(c *fiber.Ctx) error {
		return New().
			Success(true).
			Data(testData).
			Send(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	assert.Equal(t, true, result["success"])
	data := result["data"].(map[string]interface{})
	assert.Equal(t, "value", data["key"])
}

func TestBuilderWithError(t *testing.T) {
	app := setupTestApp()

	app.Get("/test", func(c *fiber.Ctx) error {
		return New().
			Success(false).
			Status(fiber.StatusBadRequest).
			Error("INVALID_INPUT", "Invalid input").
			Send(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result Response
	json.Unmarshal(body, &result)

	assert.False(t, result.Success)
	assert.NotNil(t, result.Error)
	assert.Equal(t, "Invalid input", result.Error.Message)
}

func TestBuilderWithValidationErrors(t *testing.T) {
	app := setupTestApp()

	validationErrors := []ValidationError{
		{Field: "email", Message: "Email is required"},
		{Field: "password", Message: "Password too short"},
	}

	app.Get("/test", func(c *fiber.Ctx) error {
		return New().
			Success(false).
			Status(fiber.StatusUnprocessableEntity).
			ValidationErrors(validationErrors).
			Send(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result Response
	json.Unmarshal(body, &result)

	assert.False(t, result.Success)
	assert.NotNil(t, result.Error)
	assert.Len(t, result.Error.Validation, 2)
}

func TestBuilderWithStatusCode(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		isSuccess  bool
	}{
		{"OK", http.StatusOK, true},
		{"Created", http.StatusCreated, true},
		{"BadRequest", http.StatusBadRequest, false},
		{"Unauthorized", http.StatusUnauthorized, false},
		{"NotFound", http.StatusNotFound, false},
		{"InternalServerError", http.StatusInternalServerError, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := setupTestApp()

			app.Get("/test", func(c *fiber.Ctx) error {
				return New().
					Success(tc.isSuccess).
					Status(tc.statusCode).
					Send(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, tc.statusCode, resp.StatusCode)
		})
	}
}

func TestConvenienceFunctions(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		app := setupTestApp()
		app.Get("/test", func(c *fiber.Ctx) error {
			return OK(c, map[string]string{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Created", func(t *testing.T) {
		app := setupTestApp()
		app.Get("/test", func(c *fiber.Ctx) error {
			return Created(c, map[string]string{"id": "123"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("BadRequest", func(t *testing.T) {
		app := setupTestApp()
		app.Get("/test", func(c *fiber.Ctx) error {
			return BadRequest(c, "VALIDATION_ERROR", "Invalid input")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("NotFound", func(t *testing.T) {
		app := setupTestApp()
		app.Get("/test", func(c *fiber.Ctx) error {
			return NotFound(c, "Resource not found")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
