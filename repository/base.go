package repository

import (
	"context"
	"errors"
	"reflect"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("entity not found")
	ErrAlreadyExists = errors.New("entity already exists")
	ErrInvalidID     = errors.New("invalid entity ID")
)

// BaseEntity defines the interface for entities with ID
type BaseEntity interface {
	GetID() uuid.UUID
	SetID(id uuid.UUID)
}

// BaseModel is a base model that can be embedded
type BaseModel struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt int64           `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt int64           `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"index" json:"-"`
}

// GetID returns the entity ID
func (m BaseModel) GetID() uuid.UUID {
	return m.ID
}

// SetID sets the entity ID
func (m *BaseModel) SetID(id uuid.UUID) {
	m.ID = id
}

// TenantBaseModel is a base model with tenant support
type TenantBaseModel struct {
	BaseModel
	TenantID uuid.UUID `gorm:"type:uuid;index;not null" json:"tenant_id"`
}

// ============================================
// Generic Repository Interface
// ============================================

// Repository defines the generic repository interface
type Repository[T any] interface {
	Create(ctx context.Context, entity *T) error
	CreateBatch(ctx context.Context, entities []*T) error
	Update(ctx context.Context, entity *T) error
	UpdateFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*T, error)
	FindAll(ctx context.Context) ([]T, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]T, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	Count(ctx context.Context) (int64, error)
}

// ============================================
// GORM Repository Implementation
// ============================================

// GormRepository is a generic GORM repository
type GormRepository[T any] struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM repository
func NewGormRepository[T any](db *gorm.DB) *GormRepository[T] {
	return &GormRepository[T]{db: db}
}

// DB returns the underlying GORM DB (for custom queries)
func (r *GormRepository[T]) DB() *gorm.DB {
	return r.db
}

// Create inserts a new entity
func (r *GormRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// CreateBatch inserts multiple entities
func (r *GormRepository[T]) CreateBatch(ctx context.Context, entities []*T) error {
	if len(entities) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(entities, 100).Error
}

// Update updates an existing entity
func (r *GormRepository[T]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// UpdateFields updates specific fields
func (r *GormRepository[T]) UpdateFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	var entity T
	return r.db.WithContext(ctx).Model(&entity).Where("id = ?", id).Updates(fields).Error
}

// Delete hard deletes an entity
func (r *GormRepository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	var entity T
	return r.db.WithContext(ctx).Unscoped().Delete(&entity, id).Error
}

// SoftDelete soft deletes an entity
func (r *GormRepository[T]) SoftDelete(ctx context.Context, id uuid.UUID) error {
	var entity T
	return r.db.WithContext(ctx).Delete(&entity, id).Error
}

// FindByID finds an entity by ID
func (r *GormRepository[T]) FindByID(ctx context.Context, id uuid.UUID) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &entity, err
}

// FindAll returns all entities
func (r *GormRepository[T]) FindAll(ctx context.Context) ([]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).Find(&entities).Error
	return entities, err
}

// FindByIDs finds entities by multiple IDs
func (r *GormRepository[T]) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]T, error) {
	var entities []T
	if len(ids) == 0 {
		return entities, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&entities).Error
	return entities, err
}

// Exists checks if an entity exists
func (r *GormRepository[T]) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	var entity T
	err := r.db.WithContext(ctx).Model(&entity).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// Count returns the total count of entities
func (r *GormRepository[T]) Count(ctx context.Context) (int64, error) {
	var count int64
	var entity T
	err := r.db.WithContext(ctx).Model(&entity).Count(&count).Error
	return count, err
}

// ============================================
// Query Builder
// ============================================

// Query provides a fluent query builder
type Query[T any] struct {
	db *gorm.DB
}

// NewQuery creates a new query builder
func (r *GormRepository[T]) Query() *Query[T] {
	var entity T
	return &Query[T]{db: r.db.Model(&entity)}
}

// WithContext sets the context
func (q *Query[T]) WithContext(ctx context.Context) *Query[T] {
	q.db = q.db.WithContext(ctx)
	return q
}

// Where adds a where condition
func (q *Query[T]) Where(query interface{}, args ...interface{}) *Query[T] {
	q.db = q.db.Where(query, args...)
	return q
}

// WhereNot adds a NOT where condition
func (q *Query[T]) WhereNot(query interface{}, args ...interface{}) *Query[T] {
	q.db = q.db.Not(query, args...)
	return q
}

// Or adds an OR condition
func (q *Query[T]) Or(query interface{}, args ...interface{}) *Query[T] {
	q.db = q.db.Or(query, args...)
	return q
}

// Order adds ordering
func (q *Query[T]) Order(value interface{}) *Query[T] {
	q.db = q.db.Order(value)
	return q
}

// Limit sets the limit
func (q *Query[T]) Limit(limit int) *Query[T] {
	q.db = q.db.Limit(limit)
	return q
}

// Offset sets the offset
func (q *Query[T]) Offset(offset int) *Query[T] {
	q.db = q.db.Offset(offset)
	return q
}

// Preload preloads associations
func (q *Query[T]) Preload(query string, args ...interface{}) *Query[T] {
	q.db = q.db.Preload(query, args...)
	return q
}

// Joins adds joins
func (q *Query[T]) Joins(query string, args ...interface{}) *Query[T] {
	q.db = q.db.Joins(query, args...)
	return q
}

// Select specifies fields to select
func (q *Query[T]) Select(query interface{}, args ...interface{}) *Query[T] {
	q.db = q.db.Select(query, args...)
	return q
}

// Find executes the query and returns results
func (q *Query[T]) Find() ([]T, error) {
	var entities []T
	err := q.db.Find(&entities).Error
	return entities, err
}

// First returns the first result
func (q *Query[T]) First() (*T, error) {
	var entity T
	err := q.db.First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &entity, err
}

// Count returns the count of matching records
func (q *Query[T]) Count() (int64, error) {
	var count int64
	err := q.db.Count(&count).Error
	return count, err
}

// Paginate returns paginated results
func (q *Query[T]) Paginate(page, pageSize int) ([]T, int64, error) {
	var entities []T
	var total int64

	// Clone the query for count
	countDB := q.db.Session(&gorm.Session{})
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := q.db.Offset(offset).Limit(pageSize).Find(&entities).Error
	return entities, total, err
}

// ============================================
// Tenant-Scoped Repository
// ============================================

// TenantRepository adds tenant scoping to queries
type TenantRepository[T any] struct {
	*GormRepository[T]
	tenantIDField string
}

// NewTenantRepository creates a tenant-scoped repository
func NewTenantRepository[T any](db *gorm.DB, tenantIDField string) *TenantRepository[T] {
	if tenantIDField == "" {
		tenantIDField = "tenant_id"
	}
	return &TenantRepository[T]{
		GormRepository: NewGormRepository[T](db),
		tenantIDField:  tenantIDField,
	}
}

// ForTenant returns a scoped query for a specific tenant
func (r *TenantRepository[T]) ForTenant(tenantID uuid.UUID) *Query[T] {
	return &Query[T]{
		db: r.db.Where(r.tenantIDField+" = ?", tenantID),
	}
}

// FindByIDForTenant finds an entity ensuring tenant ownership
func (r *TenantRepository[T]) FindByIDForTenant(ctx context.Context, id, tenantID uuid.UUID) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).Where("id = ? AND "+r.tenantIDField+" = ?", id, tenantID).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &entity, err
}

// ============================================
// Helper Functions
// ============================================

// GetEntityType returns the type name of an entity
func GetEntityType[T any]() string {
	var entity T
	t := reflect.TypeOf(entity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
