package audit

import (
	"context"

	"github.com/google/uuid"
)

// NoopLogger is a no-op implementation for testing
type NoopLogger struct{}

// NewNoopLogger creates a new noop logger
func NewNoopLogger() *NoopLogger {
	return &NoopLogger{}
}

// Log does nothing
func (n *NoopLogger) Log(ctx context.Context, entry *AuditLog) error {
	return nil
}

// LogAction does nothing
func (n *NoopLogger) LogAction(ctx context.Context, tenantID, userID uuid.UUID, action, entityType string, entityID *uuid.UUID, changes map[string]interface{}) error {
	return nil
}

// Query returns empty result
func (n *NoopLogger) Query(ctx context.Context, filter *Filter) ([]*AuditLog, error) {
	return []*AuditLog{}, nil
}
