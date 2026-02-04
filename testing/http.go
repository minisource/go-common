package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// ============================================
// HTTP Test Helpers
// ============================================

// TestApp creates a test Fiber app
func TestApp() *fiber.App {
	return fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})
}

// HTTPRequest represents a test HTTP request
type HTTPRequest struct {
	Method  string
	Path    string
	Body    interface{}
	Headers map[string]string
}

// HTTPResponse represents a test HTTP response
type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// DoRequest performs a test request against Fiber app
func DoRequest(app *fiber.App, req HTTPRequest) (*HTTPResponse, error) {
	var body io.Reader
	if req.Body != nil {
		jsonBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(jsonBytes)
	}

	httpReq := httptest.NewRequest(req.Method, req.Path, body)
	httpReq.Header.Set("Content-Type", "application/json")

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := app.Test(httpReq, -1)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}, nil
}

// Get performs GET request
func Get(app *fiber.App, path string, headers ...map[string]string) (*HTTPResponse, error) {
	h := make(map[string]string)
	if len(headers) > 0 {
		h = headers[0]
	}
	return DoRequest(app, HTTPRequest{
		Method:  "GET",
		Path:    path,
		Headers: h,
	})
}

// Post performs POST request
func Post(app *fiber.App, path string, body interface{}, headers ...map[string]string) (*HTTPResponse, error) {
	h := make(map[string]string)
	if len(headers) > 0 {
		h = headers[0]
	}
	return DoRequest(app, HTTPRequest{
		Method:  "POST",
		Path:    path,
		Body:    body,
		Headers: h,
	})
}

// Put performs PUT request
func Put(app *fiber.App, path string, body interface{}, headers ...map[string]string) (*HTTPResponse, error) {
	h := make(map[string]string)
	if len(headers) > 0 {
		h = headers[0]
	}
	return DoRequest(app, HTTPRequest{
		Method:  "PUT",
		Path:    path,
		Body:    body,
		Headers: h,
	})
}

// Delete performs DELETE request
func Delete(app *fiber.App, path string, headers ...map[string]string) (*HTTPResponse, error) {
	h := make(map[string]string)
	if len(headers) > 0 {
		h = headers[0]
	}
	return DoRequest(app, HTTPRequest{
		Method:  "DELETE",
		Path:    path,
		Headers: h,
	})
}

// ============================================
// Response Parsing Helpers
// ============================================

// ParseJSON parses response body as JSON
func (r *HTTPResponse) ParseJSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// BodyString returns body as string
func (r *HTTPResponse) BodyString() string {
	return string(r.Body)
}

// IsOK checks if status is 2xx
func (r *HTTPResponse) IsOK() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// ============================================
// Assertion Helpers
// ============================================

// AssertStatus asserts response status code
func AssertStatus(t *testing.T, resp *HTTPResponse, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		t.Errorf("expected status %d, got %d. Body: %s", expected, resp.StatusCode, resp.BodyString())
	}
}

// AssertOK asserts status is 200
func AssertOK(t *testing.T, resp *HTTPResponse) {
	t.Helper()
	AssertStatus(t, resp, http.StatusOK)
}

// AssertCreated asserts status is 201
func AssertCreated(t *testing.T, resp *HTTPResponse) {
	t.Helper()
	AssertStatus(t, resp, http.StatusCreated)
}

// AssertBadRequest asserts status is 400
func AssertBadRequest(t *testing.T, resp *HTTPResponse) {
	t.Helper()
	AssertStatus(t, resp, http.StatusBadRequest)
}

// AssertUnauthorized asserts status is 401
func AssertUnauthorized(t *testing.T, resp *HTTPResponse) {
	t.Helper()
	AssertStatus(t, resp, http.StatusUnauthorized)
}

// AssertForbidden asserts status is 403
func AssertForbidden(t *testing.T, resp *HTTPResponse) {
	t.Helper()
	AssertStatus(t, resp, http.StatusForbidden)
}

// AssertNotFound asserts status is 404
func AssertNotFound(t *testing.T, resp *HTTPResponse) {
	t.Helper()
	AssertStatus(t, resp, http.StatusNotFound)
}

// AssertBodyContains asserts body contains substring
func AssertBodyContains(t *testing.T, resp *HTTPResponse, substring string) {
	t.Helper()
	if !strings.Contains(resp.BodyString(), substring) {
		t.Errorf("expected body to contain '%s', got: %s", substring, resp.BodyString())
	}
}

// AssertJSONPath asserts JSON path has expected value
func AssertJSONPath(t *testing.T, resp *HTTPResponse, path string, expected interface{}) {
	t.Helper()
	var data map[string]interface{}
	if err := resp.ParseJSON(&data); err != nil {
		t.Errorf("failed to parse JSON: %v", err)
		return
	}

	parts := strings.Split(path, ".")
	var current interface{} = data
	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			t.Errorf("path '%s' not found in response", path)
			return
		}
	}

	// Compare based on type
	switch exp := expected.(type) {
	case string:
		if str, ok := current.(string); !ok || str != exp {
			t.Errorf("expected %s = '%s', got '%v'", path, exp, current)
		}
	case float64:
		if num, ok := current.(float64); !ok || num != exp {
			t.Errorf("expected %s = %v, got %v", path, exp, current)
		}
	case int:
		if num, ok := current.(float64); !ok || int(num) != exp {
			t.Errorf("expected %s = %v, got %v", path, exp, current)
		}
	case bool:
		if b, ok := current.(bool); !ok || b != exp {
			t.Errorf("expected %s = %v, got %v", path, exp, current)
		}
	default:
		t.Errorf("unsupported type for comparison")
	}
}

// ============================================
// Context Helpers
// ============================================

// TestContext creates a context with timeout
func TestContext() context.Context {
	return context.Background()
}
