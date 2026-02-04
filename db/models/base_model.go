package models

import (
    "context"
    "database/sql"
    "time"
    "github.com/google/uuid"
)

// BaseModel represents the common fields for all models
type BaseModel struct {
    ID         uuid.UUID    `json:"id"`
    CreatedAt  time.Time    `json:"created_at"`
    ModifiedAt sql.NullTime `json:"modified_at,omitempty"`
    DeletedAt  sql.NullTime `json:"deleted_at,omitempty"`
    CreatedBy  int          `json:"created_by"`
    ModifiedBy sql.NullInt64 `json:"modified_by,omitempty"`
    DeletedBy  sql.NullInt64 `json:"deleted_by,omitempty"`
}

// SetUserContext sets the user ID in the context for database operations
func SetUserContext(ctx context.Context, userID int) context.Context {
    return context.WithValue(ctx, "UserId", userID)
}

// GetUserFromContext retrieves the user ID from context
func GetUserFromContext(ctx context.Context) int {
    value := ctx.Value("UserId")
    if value == nil {
        return -1
    }
    switch v := value.(type) {
    case float64:
        return int(v)
    case int:
        return v
    default:
        return -1
    }
}