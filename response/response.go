package response

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// Response represents a standardized API response
type Response struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data,omitempty"`
	Error      *ErrorInfo  `json:"error,omitempty"`
	Meta       *Meta       `json:"meta,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	TraceID    string      `json:"traceId,omitempty"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    string            `json:"details,omitempty"`
	Field      string            `json:"field,omitempty"`
	Validation []ValidationError `json:"validation,omitempty"`
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// Meta contains response metadata
type Meta struct {
	RequestID string `json:"requestId,omitempty"`
	Version   string `json:"version,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

// Pagination contains pagination information
type Pagination struct {
	Page       int    `json:"page,omitempty"`
	PerPage    int    `json:"perPage,omitempty"`
	Total      int64  `json:"total"`
	TotalPages int    `json:"totalPages"`
	HasNext    bool   `json:"hasNext"`
	HasPrev    bool   `json:"hasPrev"`
	NextCursor string `json:"nextCursor,omitempty"`
	PrevCursor string `json:"prevCursor,omitempty"`
}

// Builder for creating responses
type Builder struct {
	response   Response
	statusCode int
}

// New creates a new response builder
func New() *Builder {
	return &Builder{
		response: Response{
			Success: true,
		},
		statusCode: http.StatusOK,
	}
}

// Success sets the success status
func (b *Builder) Success(success bool) *Builder {
	b.response.Success = success
	return b
}

// Data sets the response data
func (b *Builder) Data(data interface{}) *Builder {
	b.response.Data = data
	return b
}

// WithMeta adds metadata to the response
func (b *Builder) WithMeta(meta *Meta) *Builder {
	b.response.Meta = meta
	return b
}

// WithPagination adds pagination info
func (b *Builder) WithPagination(p *Pagination) *Builder {
	b.response.Pagination = p
	return b
}

// WithTraceID adds trace ID
func (b *Builder) WithTraceID(traceID string) *Builder {
	b.response.TraceID = traceID
	return b
}

// Status sets the HTTP status code
func (b *Builder) Status(code int) *Builder {
	b.statusCode = code
	return b
}

// Error sets error information
func (b *Builder) Error(code, message string) *Builder {
	b.response.Success = false
	b.response.Error = &ErrorInfo{
		Code:    code,
		Message: message,
	}
	return b
}

// ErrorWithDetails sets error with details
func (b *Builder) ErrorWithDetails(code, message, details string) *Builder {
	b.response.Success = false
	b.response.Error = &ErrorInfo{
		Code:    code,
		Message: message,
		Details: details,
	}
	return b
}

// ValidationErrors sets validation errors
func (b *Builder) ValidationErrors(errors []ValidationError) *Builder {
	b.response.Success = false
	b.response.Error = &ErrorInfo{
		Code:       "VALIDATION_ERROR",
		Message:    "Validation failed",
		Validation: errors,
	}
	return b
}

// Build returns the response
func (b *Builder) Build() Response {
	return b.response
}

// Send sends the response via Fiber context
func (b *Builder) Send(c *fiber.Ctx) error {
	// Extract trace ID from context if not set
	if b.response.TraceID == "" {
		if traceID, ok := c.Locals("traceId").(string); ok {
			b.response.TraceID = traceID
		}
	}
	return c.Status(b.statusCode).JSON(b.response)
}

// ============================================
// Convenience functions for common responses
// ============================================

// OK sends a successful response with data
func OK(c *fiber.Ctx, data interface{}) error {
	return New().Data(data).Send(c)
}

// OKWithPagination sends paginated data
func OKWithPagination(c *fiber.Ctx, data interface{}, pagination *Pagination) error {
	return New().Data(data).WithPagination(pagination).Send(c)
}

// Created sends a 201 response
func Created(c *fiber.Ctx, data interface{}) error {
	return New().Status(http.StatusCreated).Data(data).Send(c)
}

// NoContent sends a 204 response
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}

// BadRequest sends a 400 error
func BadRequest(c *fiber.Ctx, code, message string) error {
	return New().Status(http.StatusBadRequest).Error(code, message).Send(c)
}

// Unauthorized sends a 401 error
func Unauthorized(c *fiber.Ctx, message string) error {
	return New().Status(http.StatusUnauthorized).Error("UNAUTHORIZED", message).Send(c)
}

// Forbidden sends a 403 error
func Forbidden(c *fiber.Ctx, message string) error {
	return New().Status(http.StatusForbidden).Error("FORBIDDEN", message).Send(c)
}

// NotFound sends a 404 error
func NotFound(c *fiber.Ctx, message string) error {
	return New().Status(http.StatusNotFound).Error("NOT_FOUND", message).Send(c)
}

// Conflict sends a 409 error
func Conflict(c *fiber.Ctx, message string) error {
	return New().Status(http.StatusConflict).Error("CONFLICT", message).Send(c)
}

// UnprocessableEntity sends a 422 error with validation errors
func UnprocessableEntity(c *fiber.Ctx, errors []ValidationError) error {
	return New().Status(http.StatusUnprocessableEntity).ValidationErrors(errors).Send(c)
}

// InternalError sends a 500 error
func InternalError(c *fiber.Ctx, message string) error {
	return New().Status(http.StatusInternalServerError).Error("INTERNAL_ERROR", message).Send(c)
}

// ServiceUnavailable sends a 503 error
func ServiceUnavailable(c *fiber.Ctx, message string) error {
	return New().Status(http.StatusServiceUnavailable).Error("SERVICE_UNAVAILABLE", message).Send(c)
}

// ============================================
// Common response types for swagger docs
// ============================================

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services,omitempty"`
	Version   string            `json:"version,omitempty"`
	Uptime    string            `json:"uptime,omitempty"`
}

// SuccessMessage represents a simple success response with a message
type SuccessMessage struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ReadinessResponse represents readiness check response
type ReadinessResponse struct {
	Ready   bool   `json:"ready"`
	Message string `json:"message"`
}

// LivenessResponse represents liveness check response
type LivenessResponse struct {
	Alive bool `json:"alive"`
}

// OKMessage sends a success response with a message
func OKMessage(c *fiber.Ctx, message string) error {
	return c.JSON(SuccessMessage{
		Success: true,
		Message: message,
	})
}
