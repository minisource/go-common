package logging

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// WithContext extracts TraceID and SpanID from context and adds to logging fields
func WithContext(ctx context.Context, extra map[ExtraKey]interface{}) map[ExtraKey]interface{} {
	if extra == nil {
		extra = make(map[ExtraKey]interface{})
	}
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		extra[TraceID] = span.SpanContext().TraceID().String()
		extra[SpanID] = span.SpanContext().SpanID().String()
	}
	return extra
}
