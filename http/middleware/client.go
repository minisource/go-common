package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type instrumentedRoundTripper struct {
	next http.RoundTripper
}

// RoundTrip executes request and injects context propagation headers
func (rt *instrumentedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	tracer := otel.Tracer("http-client")
	ctx, span := tracer.Start(ctx, req.Method+" "+req.URL.Path, trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	// Inject trace headers into request
	req = req.WithContext(ctx)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Record request attributes
	span.SetAttributes(
		attribute.String("http.method", req.Method),
		attribute.String("http.url", req.URL.String()),
		attribute.String("http.host", req.URL.Host),
	)

	resp, err := rt.next.RoundTrip(req)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(
		attribute.Int("http.status_code", resp.StatusCode),
	)

	return resp, nil
}

// NewInstrumentedRoundTripper wraps a RoundTripper with HTTP tracing
func NewInstrumentedRoundTripper(next http.RoundTripper) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	return &instrumentedRoundTripper{next: next}
}

// NewInstrumentedClient creates/wraps HTTP Client with trace propagation
func NewInstrumentedClient(client *http.Client) *http.Client {
	if client == nil {
		client = &http.Client{}
	}
	client.Transport = NewInstrumentedRoundTripper(client.Transport)
	return client
}
