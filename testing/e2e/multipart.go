package e2e

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
)

// DoMultipart uploads a file via multipart/form-data.
func (c *Client) DoMultipart(method, path, fieldName, filename string, content []byte, extraFields map[string]string) (*http.Response, []byte, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile(fieldName, filename)
	if err != nil {
		return nil, nil, err
	}
	if _, err := part.Write(content); err != nil {
		return nil, nil, err
	}
	for k, v := range extraFields {
		if err := w.WriteField(k, v); err != nil {
			return nil, nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequest(method, c.BaseURL+path, &buf)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
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

// StorageServiceHeaders returns auth headers for storage API (service token + tenant/user UUIDs).
func StorageServiceHeaders(t *testing.T, authBase string) map[string]string {
	t.Helper()
	token := ServiceToken(t, authBase, "storage-service", "storage-service-secret-key")
	h := Bearer(token)
	h["X-Tenant-ID"] = "00000000-0000-0000-0000-000000000001"
	h["X-User-ID"] = "00000000-0000-0000-0000-000000000002"
	return h
}

// AdminAuthContext returns token, user ID, and a default tenant UUID for storage tests.
func AdminAuthContext(t *testing.T, authBase, email, password string) (token, userID, tenantID string, headers map[string]string) {
	t.Helper()
	token = LoginAuth(t, authBase, email, password)
	auth := NewClient(authBase, Bearer(token))
	resp, body, err := auth.Do(http.MethodGet, "/api/v1/users/me", nil)
	if err != nil {
		t.Fatal(err)
	}
	ExpectStatus(t, resp, body, http.StatusOK)
	var parsed map[string]any
	ParseJSON(t, body, &parsed)
	userID = GetString(parsed, "data", "id")
	if userID == "" {
		userID = GetString(parsed, "id")
	}
	if userID == "" {
		t.Fatal("no user id from /users/me")
	}
	tenantID = GetString(parsed, "data", "tenantId")
	if tenantID == "" {
		tenantID = "00000000-0000-0000-0000-000000000001"
	}
	h := Bearer(token)
	h["X-Tenant-ID"] = tenantID
	h["X-User-ID"] = userID
	return token, userID, tenantID, h
}

// ExtractID tries common response shapes for created resource IDs.
func ExtractID(t *testing.T, body []byte) string {
	t.Helper()
	var parsed map[string]any
	ParseJSON(t, body, &parsed)
	for _, keys := range [][]string{
		{"data", "id"},
		{"data", "_id"},
		{"result", "id"},
		{"id"},
	} {
		if id := GetString(parsed, keys...); id != "" {
			return id
		}
	}
	if data, ok := parsed["data"].(map[string]any); ok {
		if id, ok := data["id"].(string); ok && id != "" {
			return id
		}
	}
	t.Fatalf("no id in response: %s", truncate(body, 400))
	return ""
}
