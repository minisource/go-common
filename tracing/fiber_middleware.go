package tracing

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// MiddlewareConfig holds configuration for tracing middleware
type MiddlewareConfig struct {
	ServiceName   string
	SkipPaths     []string
	TracerName    string
	SpanNameFunc  func(*fiber.Ctx) string
	RecordBody    bool
	RecordHeaders bool
}

// DefaultMiddlewareConfig returns default middleware configuration
func DefaultMiddlewareConfig() MiddlewareConfig {
	return MiddlewareConfig{
		ServiceName: "fiber-app",
		SkipPaths:   []string{"/health", "/ready", "/healthz", "/readyz", "/metrics"},
		TracerName:  "fiber-middleware",
		SpanNameFunc: func(c *fiber.Ctx) string {
			return fmt.Sprintf("%s %s", c.Method(), c.Path())
		},
		RecordBody:    false,
		RecordHeaders: false,
	}
}

// Middleware creates a Fiber middleware for OpenTelemetry tracing
func Middleware(cfg MiddlewareConfig) fiber.Handler {
	tracer := otel.Tracer(cfg.TracerName)
	propagator := otel.GetTextMapPropagator()

	return func(c *fiber.Ctx) error {
		// Check if path should be skipped
		path := c.Path()
		for _, skipPath := range cfg.SkipPaths {
			if path == skipPath {
				return c.Next()
			}
		}

		// Extract context from incoming request headers
		ctx := propagator.Extract(c.Context(), &headerCarrier{ctx: c})

		// Start span
		spanName := cfg.SpanNameFunc(c)
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethod(c.Method()),
				semconv.HTTPRoute(c.Route().Path),
				semconv.HTTPURL(c.OriginalURL()),
				semconv.HTTPScheme(c.Protocol()),
				semconv.NetHostName(c.Hostname()),
				attribute.String("http.client_ip", c.IP()),
				attribute.String("http.user_agent", c.Get("User-Agent")),
			),
		)
		defer span.End()

		// Add trace ID to response header
		if span.SpanContext().HasTraceID() {
			c.Set("X-Trace-ID", span.SpanContext().TraceID().String())
		}

		// Store context with span in Fiber context
		c.SetUserContext(ctx)
		c.Locals("traceId", span.SpanContext().TraceID().String())
		c.Locals("spanId", span.SpanContext().SpanID().String())

		// Record request body if enabled
		if cfg.RecordBody && len(c.Body()) > 0 && len(c.Body()) < 10000 {
			span.SetAttributes(attribute.String("http.request_body", string(c.Body())))
		}

		// Process request
		err := c.Next()

		// Record response status
		statusCode := c.Response().StatusCode()
		span.SetAttributes(semconv.HTTPStatusCode(statusCode))

		// Record error if any
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("error.message", err.Error()))
		}

		// Mark span as error if status code >= 400
		if statusCode >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
		}

		return err
	}
}

// headerCarrier adapts Fiber context for OpenTelemetry propagation
type headerCarrier struct {
	ctx *fiber.Ctx
}

func (c *headerCarrier) Get(key string) string {
	return c.ctx.Get(key)
}

func (c *headerCarrier) Set(key, value string) {
	c.ctx.Set(key, value)
}

func (c *headerCarrier) Keys() []string {
	keys := make([]string, 0)
	c.ctx.Request().Header.VisitAll(func(k, v []byte) {
		keys = append(keys, string(k))
	})
	return keys
}

// InjectHeaders injects trace context into outgoing request headers
func InjectHeaders(ctx *fiber.Ctx, headers map[string]string) {
	propagator := otel.GetTextMapPropagator()
	carrier := &mapCarrier{headers: headers}
	propagator.Inject(ctx.UserContext(), carrier)
}

type mapCarrier struct {
	headers map[string]string
}

func (c *mapCarrier) Get(key string) string {
	return c.headers[key]
}

func (c *mapCarrier) Set(key, value string) {
	c.headers[key] = value
}

func (c *mapCarrier) Keys() []string {
	keys := make([]string, 0, len(c.headers))
	for k := range c.headers {
		keys = append(keys, k)
	}
	return keys
}
