package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/minisource/go-common/logging"
)

// Client is a reusable HTTP client with retry, logging, and error handling
type Client struct {
	httpClient   *http.Client
	logger       logging.Logger
	retryConfig  RetryConfig
	baseURL      string
	serviceName  string
	interceptors []Interceptor
}

// Config holds HTTP client configuration
type Config struct {
	BaseURL      string
	ServiceName  string
	Timeout      time.Duration
	RetryConfig  RetryConfig
	Logger       logging.Logger
	Interceptors []Interceptor
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors []int // HTTP status codes to retry
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		InitialDelay:  500 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []int{
			http.StatusRequestTimeout,
			http.StatusTooManyRequests,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		},
	}
}

// Interceptor allows modifying requests before they are sent
type Interceptor func(ctx context.Context, req *http.Request) error

// NewClient creates a new HTTP client
func NewClient(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	if cfg.RetryConfig.MaxRetries == 0 {
		cfg.RetryConfig = DefaultRetryConfig()
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger:       cfg.Logger,
		retryConfig:  cfg.RetryConfig,
		baseURL:      cfg.BaseURL,
		serviceName:  cfg.ServiceName,
		interceptors: cfg.Interceptors,
	}
}

// Request represents an HTTP request
type Request struct {
	Method  string
	Path    string
	Body    interface{}
	Headers map[string]string
	Query   map[string]string
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// Do executes an HTTP request with retry logic and logging
func (c *Client) Do(ctx context.Context, req Request) (*Response, error) {
	startTime := time.Now()

	c.logger.Debug(logging.General, logging.ExternalService, "Starting HTTP request", map[logging.ExtraKey]interface{}{
		"service": c.serviceName,
		"method":  req.Method,
		"path":    req.Path,
	})

	var lastErr error
	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := c.calculateBackoff(attempt)
			c.logger.Debug(logging.General, logging.ExternalService, "Retrying request", map[logging.ExtraKey]interface{}{
				"service": c.serviceName,
				"attempt": attempt,
				"delay":   delay.String(),
			})

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		resp, err := c.doRequest(ctx, req, attempt)
		if err == nil && !c.shouldRetry(resp.StatusCode) {
			duration := time.Since(startTime)
			c.logger.Info(logging.General, logging.ExternalService, "HTTP request completed", map[logging.ExtraKey]interface{}{
				"service":    c.serviceName,
				"method":     req.Method,
				"path":       req.Path,
				"statusCode": resp.StatusCode,
				"duration":   duration.String(),
				"attempt":    attempt + 1,
			})
			return resp, nil
		}

		if err != nil {
			lastErr = err
			c.logger.Warn(logging.General, logging.ExternalService, "HTTP request failed", map[logging.ExtraKey]interface{}{
				"service": c.serviceName,
				"method":  req.Method,
				"path":    req.Path,
				"attempt": attempt + 1,
				"error":   err.Error(),
			})
		} else if c.shouldRetry(resp.StatusCode) {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(resp.Body))
			c.logger.Warn(logging.General, logging.ExternalService, "HTTP request returned retryable error", map[logging.ExtraKey]interface{}{
				"service":    c.serviceName,
				"method":     req.Method,
				"path":       req.Path,
				"statusCode": resp.StatusCode,
				"attempt":    attempt + 1,
			})
		}
	}

	duration := time.Since(startTime)
	c.logger.Error(logging.General, logging.ExternalService, "HTTP request failed after retries", map[logging.ExtraKey]interface{}{
		"service":  c.serviceName,
		"method":   req.Method,
		"path":     req.Path,
		"attempts": c.retryConfig.MaxRetries + 1,
		"duration": duration.String(),
		"error":    lastErr.Error(),
	})

	return nil, NewServiceUnavailableError(c.serviceName, lastErr)
}

func (c *Client) doRequest(ctx context.Context, req Request, attempt int) (*Response, error) {
	url := c.baseURL + req.Path
	if len(req.Query) > 0 {
		url += "?"
		first := true
		for k, v := range req.Query {
			if !first {
				url += "&"
			}
			url += fmt.Sprintf("%s=%s", k, v)
			first = false
		}
	}

	var bodyReader io.Reader
	if req.Body != nil {
		jsonBody, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)

		c.logger.Debug(logging.General, logging.ExternalService, "Request body", map[logging.ExtraKey]interface{}{
			"service": c.serviceName,
			"body":    string(jsonBody),
			"attempt": attempt + 1,
		})
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Add custom headers
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Run interceptors
	for _, interceptor := range c.interceptors {
		if err := interceptor(ctx, httpReq); err != nil {
			return nil, fmt.Errorf("interceptor failed: %w", err)
		}
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	c.logger.Debug(logging.General, logging.ExternalService, "Response received", map[logging.ExtraKey]interface{}{
		"service":    c.serviceName,
		"statusCode": httpResp.StatusCode,
		"body":       string(body),
		"attempt":    attempt + 1,
	})

	return &Response{
		StatusCode: httpResp.StatusCode,
		Body:       body,
		Headers:    httpResp.Header,
	}, nil
}

func (c *Client) calculateBackoff(attempt int) time.Duration {
	delay := float64(c.retryConfig.InitialDelay) * pow(c.retryConfig.BackoffFactor, float64(attempt-1))
	if delay > float64(c.retryConfig.MaxDelay) {
		return c.retryConfig.MaxDelay
	}
	return time.Duration(delay)
}

func (c *Client) shouldRetry(statusCode int) bool {
	for _, code := range c.retryConfig.RetryableErrors {
		if code == statusCode {
			return true
		}
	}
	return false
}

func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

// Get is a convenience method for GET requests
func (c *Client) Get(ctx context.Context, path string, headers map[string]string) (*Response, error) {
	return c.Do(ctx, Request{
		Method:  http.MethodGet,
		Path:    path,
		Headers: headers,
	})
}

// Post is a convenience method for POST requests
func (c *Client) Post(ctx context.Context, path string, body interface{}, headers map[string]string) (*Response, error) {
	return c.Do(ctx, Request{
		Method:  http.MethodPost,
		Path:    path,
		Body:    body,
		Headers: headers,
	})
}

// Put is a convenience method for PUT requests
func (c *Client) Put(ctx context.Context, path string, body interface{}, headers map[string]string) (*Response, error) {
	return c.Do(ctx, Request{
		Method:  http.MethodPut,
		Path:    path,
		Body:    body,
		Headers: headers,
	})
}

// Delete is a convenience method for DELETE requests
func (c *Client) Delete(ctx context.Context, path string, headers map[string]string) (*Response, error) {
	return c.Do(ctx, Request{
		Method:  http.MethodDelete,
		Path:    path,
		Headers: headers,
	})
}

// DecodeJSON decodes JSON response into target
func (r *Response) DecodeJSON(target interface{}) error {
	if err := json.Unmarshal(r.Body, target); err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}
	return nil
}
