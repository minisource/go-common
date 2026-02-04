package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TracingConfig holds configuration for tracing middleware
type TracingConfig struct {
	ServiceName string
	Tracer      trace.Tracer
}

// Tracing creates OpenTelemetry tracing middleware for Fiber
func Tracing(config TracingConfig) fiber.Handler {
	if config.Tracer == nil {
		config.Tracer = otel.Tracer(config.ServiceName)
	}

	// Use default propagator if not set
	propagator := otel.GetTextMapPropagator()
	if propagator == nil {
		propagator = propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)
	}

	return func(c *fiber.Ctx) error {
		// Extract context from headers (for distributed tracing)
		ctx := propagator.Extract(c.UserContext(), &FiberCarrier{ctx: c})

		// Start span
		spanName := c.Method() + " " + c.Route().Path
		ctx, span := config.Tracer.Start(ctx, spanName)
		defer span.End()

		// Set span attributes
		span.SetAttributes(
			attribute.String("http.method", c.Method()),
			attribute.String("http.url", string(c.Request().URI().FullURI())),
			attribute.String("http.scheme", c.Protocol()),
			attribute.String("http.host", c.Hostname()),
			attribute.String("http.user_agent", string(c.Request().Header.UserAgent())),
			attribute.String("http.route", c.Route().Path),
		)

		// Store context with span
		c.SetUserContext(ctx)

		// Call next handler
		err := c.Next()

		// Set response attributes
		statusCode := c.Response().StatusCode()
		span.SetAttributes(
			attribute.Int("http.status_code", statusCode),
			attribute.Int64("http.response_size", int64(len(c.Response().Body()))),
		)

		// Record error if any
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else if statusCode >= 400 {
			span.SetStatus(codes.Error, http.StatusText(statusCode))
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}

// FiberCarrier adapts Fiber context for OpenTelemetry propagation
type FiberCarrier struct {
	ctx *fiber.Ctx
}

// Get retrieves a value from the carrier
func (fc *FiberCarrier) Get(key string) string {
	return fc.ctx.Get(key)
}

// Set stores a value in the carrier
func (fc *FiberCarrier) Set(key string, value string) {
	fc.ctx.Set(key, value)
}

// Keys returns all keys in the carrier
func (fc *FiberCarrier) Keys() []string {
	keys := make([]string, 0)
	fc.ctx.Request().Header.VisitAll(func(key, _ []byte) {
		keys = append(keys, string(key))
	})
	return keys
}
