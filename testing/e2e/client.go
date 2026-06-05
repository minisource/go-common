// Package e2e provides HTTP helpers for end-to-end API tests against running services.
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// Client performs HTTP requests against a service base URL.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Headers    map[string]string
}

// NewClient creates a client with optional default headers.
func NewClient(baseURL string, headers map[string]string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Headers: headers,
	}
}

// BaseURLFromEnv returns env value or default.
func BaseURLFromEnv(envKey, defaultURL string) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return defaultURL
}

// RequireUp skips the test if the service health endpoint is unreachable.
func (c *Client) RequireUp(t *testing.T, healthPath string) {
	t.Helper()
	if healthPath == "" {
		healthPath = "/health"
	}
	resp, err := c.HTTPClient.Get(c.BaseURL + healthPath)
	if err != nil {
		t.Skipf("service not running at %s: %v", c.BaseURL, err)
	}
	resp.Body.Close()
	if resp.StatusCode >= 500 {
		t.Skipf("service unhealthy at %s%s: status %d", c.BaseURL, healthPath, resp.StatusCode)
	}
}

// WithHeaders returns a copy of the client with merged headers.
func (c *Client) WithHeaders(h map[string]string) *Client {
	merged := make(map[string]string, len(c.Headers)+len(h))
	for k, v := range c.Headers {
		merged[k] = v
	}
	for k, v := range h {
		merged[k] = v
	}
	return &Client{BaseURL: c.BaseURL, HTTPClient: c.HTTPClient, Headers: merged}
}

// Do sends an HTTP request.
func (c *Client) Do(method, path string, body any) (*http.Response, []byte, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.BaseURL+path, reader)
	if err != nil {
		return nil, nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, err
	}
	return resp, data, nil
}

// ExpectStatus fails the test if status is not in allowed list.
func ExpectStatus(t *testing.T, resp *http.Response, body []byte, allowed ...int) {
	t.Helper()
	for _, code := range allowed {
		if resp.StatusCode == code {
			return
		}
	}
	t.Fatalf("unexpected status %d, want one of %v, body: %s", resp.StatusCode, allowed, truncate(body, 500))
}

// ParseJSON unmarshals body into v.
func ParseJSON(t *testing.T, body []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(body, v); err != nil {
		t.Fatalf("decode json: %v, body: %s", err, truncate(body, 300))
	}
}

// GetString tries common JSON paths for nested API responses.
func GetString(data map[string]any, keys ...string) string {
	cur := any(data)
	for _, k := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur = m[k]
	}
	s, _ := cur.(string)
	return s
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}

// Case describes one API test case.
type Case struct {
	Name     string
	Method   string
	Path     string
	Body     any
	Headers  map[string]string
	WantCode []int
	Skip     bool
	SkipMsg  string
}

// RunCases executes table-driven HTTP cases.
func (c *Client) RunCases(t *testing.T, cases []Case) {
	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Skip {
				t.Skip(tc.SkipMsg)
			}
			client := c
			if len(tc.Headers) > 0 {
				client = c.WithHeaders(tc.Headers)
			}
			resp, body, err := client.Do(tc.Method, tc.Path, tc.Body)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			ExpectStatus(t, resp, body, tc.WantCode...)
		})
	}
}

// LoginAuth returns access token from auth service login.
func LoginAuth(t *testing.T, authBase, email, password string) string {
	t.Helper()
	c := NewClient(authBase, nil)
	resp, body, err := c.Do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		t.Fatalf("login request: %v", err)
	}
	ExpectStatus(t, resp, body, http.StatusOK)
	var parsed map[string]any
	ParseJSON(t, body, &parsed)
	token := GetString(parsed, "data", "accessToken")
	if token == "" {
		token = GetString(parsed, "accessToken")
	}
	if token == "" {
		t.Fatalf("no accessToken in login response: %s", string(body))
	}
	return token
}

// ServiceToken returns a service access token from auth service auth endpoint.
func ServiceToken(t *testing.T, authBase, clientID, clientSecret string) string {
	t.Helper()
	c := NewClient(authBase, nil)
	resp, body, err := c.Do(http.MethodPost, "/api/v1/service/auth", map[string]string{
		"clientId":     clientID,
		"clientSecret": clientSecret,
	})
	if err != nil {
		t.Fatalf("service auth: %v", err)
	}
	ExpectStatus(t, resp, body, http.StatusOK)
	var parsed map[string]any
	ParseJSON(t, body, &parsed)
	token := GetString(parsed, "accessToken")
	if token == "" {
		token = GetString(parsed, "data", "accessToken")
	}
	if token == "" {
		t.Fatalf("no accessToken in service auth: %s", string(body))
	}
	return token
}

// Bearer returns Authorization header value.
func Bearer(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

// TenantHeader returns X-Tenant-ID header.
func TenantHeader(id string) map[string]string {
	return map[string]string{"X-Tenant-ID": id}
}

// UniqueEmail generates a unique test email.
func UniqueEmail(prefix string) string {
	return fmt.Sprintf("%s_%d@e2e.test", prefix, time.Now().UnixNano())
}
