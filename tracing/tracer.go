package tracing

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config holds tracing configuration
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	CollectorURL   string  // OTLP collector URL (e.g., "localhost:4317")
	SamplingRate   float64 // 0.0 to 1.0
	Enabled        bool
}

// DefaultConfig returns default tracing configuration
func DefaultConfig() Config {
	return Config{
		ServiceName:    "unknown-service",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		CollectorURL:   "localhost:4317",
		SamplingRate:   1.0,
		Enabled:        false,
	}
}

// LoadConfigFromEnv loads tracing config from environment variables
func LoadConfigFromEnv() Config {
	cfg := DefaultConfig()

	if enabledStr, exists := os.LookupEnv("OTEL_ENABLED"); exists {
		if val, err := strconv.ParseBool(enabledStr); err == nil {
			cfg.Enabled = val
		}
	} else if enabledStr, exists := os.LookupEnv("TRACING_ENABLED"); exists {
		if val, err := strconv.ParseBool(enabledStr); err == nil {
			cfg.Enabled = val
		}
	}

	if svcName, exists := os.LookupEnv("OTEL_SERVICE_NAME"); exists {
		cfg.ServiceName = svcName
	} else if svcName, exists := os.LookupEnv("TRACING_SERVICE_NAME"); exists {
		cfg.ServiceName = svcName
	}

	if endpoint, exists := os.LookupEnv("OTEL_EXPORTER_OTLP_ENDPOINT"); exists {
		cfg.CollectorURL = sanitizeEndpoint(endpoint)
	} else if endpoint, exists := os.LookupEnv("JAEGER_URL"); exists {
		cfg.CollectorURL = sanitizeEndpoint(endpoint)
	}

	if samplerArg, exists := os.LookupEnv("OTEL_TRACES_SAMPLER_ARG"); exists {
		if rate, err := strconv.ParseFloat(samplerArg, 64); err == nil {
			cfg.SamplingRate = rate
		}
	}

	if env, exists := os.LookupEnv("APP_ENV"); exists {
		cfg.Environment = env
	}

	return cfg
}

func sanitizeEndpoint(endpoint string) string {
	if len(endpoint) > 7 && endpoint[:7] == "http://" {
		return endpoint[7:]
	}
	if len(endpoint) > 8 && endpoint[:8] == "https://" {
		return endpoint[8:]
	}
	return endpoint
}

// Tracer wraps OpenTelemetry tracer
type Tracer struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
	cfg      Config
}

// InitTracer initializes OpenTelemetry tracing
func InitTracer(ctx context.Context, cfg Config) (*Tracer, error) {
	if !cfg.Enabled {
		// Return a no-op tracer
		return &Tracer{
			tracer: otel.Tracer(cfg.ServiceName),
			cfg:    cfg,
		}, nil
	}

	// Create OTLP exporter without blocking dial to fail safe if collector is down
	conn, err := grpc.DialContext(ctx, cfg.CollectorURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial collector: %w", err)
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// Create resource with service information.
	// resource.Default() carries the SDK's own schema URL; merging a resource
	// built with a different semconv schema URL fails with a schema conflict,
	// so we reuse the default resource's schema URL for our attributes.
	defRes := resource.Default()
	res, err := resource.Merge(
		defRes,
		resource.NewWithAttributes(
			defRes.SchemaURL(),
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			attribute.String("environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create sampler
	var sampler sdktrace.Sampler
	if cfg.SamplingRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if cfg.SamplingRate <= 0.0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(cfg.SamplingRate)
	}

	// Create trace provider
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global provider and propagator
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Tracer{
		provider: provider,
		tracer:   provider.Tracer(cfg.ServiceName),
		cfg:      cfg,
	}, nil
}

// Shutdown gracefully shuts down the tracer
func (t *Tracer) Shutdown(ctx context.Context) error {
	if t.provider != nil {
		return t.provider.Shutdown(ctx)
	}
	return nil
}

// StartSpan starts a new span
func (t *Tracer) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

// SpanFromContext returns the span from context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddEvent adds an event to the current span
func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetAttributes sets attributes on the current span
func SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// RecordError records an error on the current span
func RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}

// GetTraceID returns the trace ID from context
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID returns the span ID from context
func GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasSpanID() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}
