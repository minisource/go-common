package e2e

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
)

// SwaggerOp is one operation from OpenAPI paths.
type SwaggerOp struct {
	Method string
	Path   string
}

// LoadSwaggerOps parses swagger.json paths (methods: get, post, ...).
func LoadSwaggerOps(doc []byte) ([]SwaggerOp, error) {
	var spec struct {
		Paths map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(doc, &spec); err != nil {
		return nil, err
	}
	var ops []SwaggerOp
	for path, raw := range spec.Paths {
		var methods map[string]json.RawMessage
		if err := json.Unmarshal(raw, &methods); err != nil {
			continue
		}
		for m := range methods {
			ops = append(ops, SwaggerOp{
				Method: strings.ToUpper(m),
				Path:   path,
			})
		}
	}
	return ops, nil
}

// LoadSwaggerOpsFromFile reads swagger.json from disk.
func LoadSwaggerOpsFromFile(path string) ([]SwaggerOp, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadSwaggerOps(b)
}

var pathParamRe = regexp.MustCompile(`\{[^}]+\}`)

// FillSwaggerPath replaces path params with placeholder IDs.
func FillSwaggerPath(path string) string {
	p := pathParamRe.ReplaceAllString(path, "00000000-0000-0000-0000-000000000001")
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	// Auth swagger paths omit /api/v1 prefix in some specs
	if !strings.HasPrefix(p, "/api/") && !strings.HasPrefix(p, "/health") && !strings.HasPrefix(p, "/ready") && !strings.HasPrefix(p, "/metrics") {
		if strings.HasPrefix(p, "/auth/") || strings.HasPrefix(p, "/users/") || strings.HasPrefix(p, "/admin/") || strings.HasPrefix(p, "/service/") || strings.HasPrefix(p, "/tokens/") {
			p = "/api/v1" + p
		}
	}
	return p
}

// RunSwaggerSmoke hits each operation; skips non-GET unless allowWrite is true.
func RunSwaggerSmoke(t *testing.T, c *Client, docPath string, headers map[string]string, allowWrite bool) {
	t.Helper()
	ops, err := LoadSwaggerOpsFromFile(docPath)
	if err != nil {
		t.Skipf("swagger doc not found: %v", err)
	}
	okStatuses := []int{
		http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent,
		http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden,
		http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity,
		http.StatusTooManyRequests, http.StatusInternalServerError,
	}
	for _, op := range ops {
		op := op
		if op.Method != http.MethodGet && !allowWrite {
			continue
		}
		if op.Method != http.MethodGet && op.Method != http.MethodPost && op.Method != http.MethodPut && op.Method != http.MethodPatch && op.Method != http.MethodDelete {
			continue
		}
		name := op.Method + "_" + strings.ReplaceAll(strings.Trim(op.Path, "/"), "/", "_")
		t.Run(name, func(t *testing.T) {
			path := FillSwaggerPath(op.Path)
			client := c
			if len(headers) > 0 {
				client = c.WithHeaders(headers)
			}
			var body any
			if op.Method == http.MethodPost || op.Method == http.MethodPut || op.Method == http.MethodPatch {
				body = map[string]any{}
			}
			resp, respBody, err := client.Do(op.Method, path, body)
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			ExpectStatus(t, resp, respBody, okStatuses...)
		})
	}
}
