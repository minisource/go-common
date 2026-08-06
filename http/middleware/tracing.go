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

// TracingConfig holds configuration for tracing middleware.
type TracingConfig struct {
	ServiceName string
	Tracer      trace.Tracer
}

// Semantic attribute keys (OTel conventions, kept as literals so every
// vendored copy of go-common compiles against its own semconv version).
const (
	attrHTTPRequestMethod  = "http.request.method"
	attrHTTPRoute          = "http.route"
	attrURLPath            = "url.path"
	attrHTTPStatusCode     = "http.response.status_code"
	attrServerAddress      = "server.address"
	attrServerPort         = "server.port"
	attrClientAddress      = "client.address"
	attrUserAgentOriginal  = "user_agent.original"
	attrNetworkProtoVer    = "network.protocol.version"
	attrErrorType          = "error.type"
	attrRequestID          = "request.id"
	attrServiceName        = "service.name"
)

// RouteTemplate returns the matched route template (e.g. "/v1/users/:id")
// after route resolution. Returns "" when no concrete route matched
// (404 fallback), so callers can fall back to the raw path.
func RouteTemplate(c *fiber.Ctx) string {
	route := c.Route()
	if route == nil || route.Path == "" || route.Path == "/*" {
		return ""
	}
	return route.Path
}

// Tracing creates OpenTelemetry tracing middleware for Fiber.
//
// The span is created before route resolution (so the initial name is generic),
// then renamed to the matched route template once the handler chain returns.
// Raw IDs and UUIDs never enter span names or metric labels.
func Tracing(config TracingConfig) fiber.Handler {
	if config.Tracer == nil {
		config.Tracer = otel.Tracer(config.ServiceName)
	}

	propagator := otel.GetTextMapPropagator()
	if propagator == nil {
		propagator = propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)
	}

	return func(c *fiber.Ctx) error {
		// Extract incoming W3C context (traceparent/tracestate/baggage).
		ctx := propagator.Extract(c.UserContext(), &FiberCarrier{ctx: c})

		method := c.Method()
		initialName := method + " " + c.Path()
		ctx, span := config.Tracer.Start(ctx, initialName)
		defer span.End()

		reqID := GetRequestID(c)

		span.SetAttributes(
			attribute.String(attrHTTPRequestMethod, method),
			attribute.String(attrURLPath, c.Path()),
			attribute.String(attrUserAgentOriginal, string(c.Request().Header.UserAgent())),
			attribute.String(attrServerAddress, c.Hostname()),
			attribute.String(attrNetworkProtoVer, string(c.Request().Header.Protocol())),
			attribute.String(attrServiceName, config.ServiceName),
		)
		if port := c.Port(); port != "" {
			span.SetAttributes(attribute.String(attrServerPort, port))
		}
		if clientIP := c.IP(); clientIP != "" && clientIP != "0.0.0.0" {
			span.SetAttributes(attribute.String(attrClientAddress, clientIP))
		}
		if reqID != "" {
			span.SetAttributes(attribute.String(attrRequestID, reqID))
		}

		// Store context with span for downstream middleware/handlers.
		c.SetUserContext(ctx)

		// Expose trace ID to clients for support correlation.
		if sc := span.SpanContext(); sc.HasTraceID() {
			c.Set("X-Trace-ID", sc.TraceID().String())
		}

		err := c.Next()

		statusCode := c.Response().StatusCode()

		// Route is now resolved: rename the span and fix http.route.
		if route := RouteTemplate(c); route != "" {
			span.SetName(method + " " + route)
			span.SetAttributes(attribute.String(attrHTTPRoute, route))
		} else {
			span.SetAttributes(attribute.String(attrHTTPRoute, c.Path()))
		}

		span.SetAttributes(attribute.Int(attrHTTPStatusCode, statusCode))

		switch {
		case err != nil:
			// Handler failure: mark the span errored without leaking the raw
			// error payload (it may contain request data).
			span.SetStatus(codes.Error, "handler error")
			span.SetAttributes(attribute.String(attrErrorType, errorCategory(err, statusCode)))
		case statusCode >= 500:
			span.SetStatus(codes.Error, http.StatusText(statusCode))
			span.SetAttributes(attribute.String(attrErrorType, http.StatusText(statusCode)))
		default:
			// 2xx/3xx/4xx are recorded with status_code; 4xx is not an error.
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}

// errorCategory returns a safe, non-payload error classification.
func errorCategory(err error, statusCode int) string {
	if statusCode >= 400 {
		return http.StatusText(statusCode)
	}
	return "handler_error"
}

// FiberCarrier adapts Fiber context for OpenTelemetry propagation.
type FiberCarrier struct {
	ctx *fiber.Ctx
}

// Get retrieves a value from the carrier.
func (fc *FiberCarrier) Get(key string) string {
	return fc.ctx.Get(key)
}

// Set stores a value in the carrier.
func (fc *FiberCarrier) Set(key string, value string) {
	fc.ctx.Set(key, value)
}

// Keys returns all keys in the carrier.
func (fc *FiberCarrier) Keys() []string {
	keys := make([]string, 0)
	fc.ctx.Request().Header.VisitAll(func(key, _ []byte) {
		keys = append(keys, string(key))
	})
	return keys
}
