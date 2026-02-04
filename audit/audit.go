package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Action types for audit logging
const (
	ActionCreate = "CREATE"
	ActionUpdate = "UPDATE"
	ActionDelete = "DELETE"
	ActionLogin  = "LOGIN"
	ActionLogout = "LOGOUT"
	ActionView   = "VIEW"
	ActionExport = "EXPORT"
)

// Entity types
const (
	EntityUser       = "USER"
	EntityRole       = "ROLE"
	EntityPermission = "PERMISSION"
	EntitySession    = "SESSION"
	EntitySetting    = "SETTING"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID         uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID   uuid.UUID              `json:"tenant_id" gorm:"type:uuid;not null;index"`
	UserID     *uuid.UUID             `json:"user_id,omitempty" gorm:"type:uuid;index"`
	Action     string                 `json:"action" gorm:"size:100;not null;index"`
	EntityType string                 `json:"entity_type" gorm:"size:100;not null;index"`
	EntityID   *uuid.UUID             `json:"entity_id,omitempty" gorm:"type:uuid;index"`
	OldValues  map[string]interface{} `json:"old_values,omitempty" gorm:"type:jsonb"`
	NewValues  map[string]interface{} `json:"new_values,omitempty" gorm:"type:jsonb"`
	IPAddress  string                 `json:"ip_address,omitempty" gorm:"size:45"`
	UserAgent  string                 `json:"user_agent,omitempty" gorm:"type:text"`
	Metadata   map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt  time.Time              `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_audit_created"`
}

// TableName overrides the table name
func (AuditLog) TableName() string {
	return "audit_logs"
}

// Logger interface for audit logging
type Logger interface {
	Log(ctx context.Context, entry *AuditLog) error
	LogAction(ctx context.Context, tenantID, userID uuid.UUID, action, entityType string, entityID *uuid.UUID, changes map[string]interface{}) error
	Query(ctx context.Context, filter *Filter) ([]*AuditLog, error)
}

// Filter for querying audit logs
type Filter struct {
	TenantID   uuid.UUID
	UserID     *uuid.UUID
	Action     string
	EntityType string
	EntityID   *uuid.UUID
	StartDate  *time.Time
	EndDate    *time.Time
	Limit      int
	Offset     int
}

// Service implements audit logging
type Service struct {
	db *gorm.DB
}

// NewService creates a new audit service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// Log writes an audit log entry
func (s *Service) Log(ctx context.Context, entry *AuditLog) error {
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	return s.db.WithContext(ctx).Create(entry).Error
}

// LogAction is a convenience method for logging actions
func (s *Service) LogAction(ctx context.Context, tenantID, userID uuid.UUID, action, entityType string, entityID *uuid.UUID, changes map[string]interface{}) error {
	entry := &AuditLog{
		TenantID:   tenantID,
		UserID:     &userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		NewValues:  changes,
	}
	return s.Log(ctx, entry)
}

// Query retrieves audit logs based on filter
func (s *Service) Query(ctx context.Context, filter *Filter) ([]*AuditLog, error) {
	query := s.db.WithContext(ctx).Model(&AuditLog{})

	query = query.Where("tenant_id = ?", filter.TenantID)

	if filter.UserID != nil {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}

	if filter.EntityType != "" {
		query = query.Where("entity_type = ?", filter.EntityType)
	}

	if filter.EntityID != nil {
		query = query.Where("entity_id = ?", filter.EntityID)
	}

	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", filter.StartDate)
	}

	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", filter.EndDate)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var logs []*AuditLog
	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, err
}

// CompareChanges creates a change map for auditing
func CompareChanges(old, new interface{}) (map[string]interface{}, error) {
	changes := make(map[string]interface{})

	oldJSON, err := json.Marshal(old)
	if err != nil {
		return nil, err
	}

	newJSON, err := json.Marshal(new)
	if err != nil {
		return nil, err
	}

	var oldMap, newMap map[string]interface{}
	if err := json.Unmarshal(oldJSON, &oldMap); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(newJSON, &newMap); err != nil {
		return nil, err
	}

	for key, newValue := range newMap {
		oldValue, exists := oldMap[key]
		if !exists || oldValue != newValue {
			changes[key] = newValue
		}
	}

	return changes, nil
}
