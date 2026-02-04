package testing

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ============================================
// Mock Repository
// ============================================

// MockEntity is a generic entity for testing
type MockEntity struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// MockRepository is an in-memory repository for testing
type MockRepository[T any] struct {
	mu    sync.RWMutex
	items map[uuid.UUID]T
	getID func(T) uuid.UUID
	setID func(*T, uuid.UUID)

	// Error injection
	ErrCreate error
	ErrUpdate error
	ErrDelete error
	ErrFind   error
	ErrList   error
}

// NewMockRepository creates a new mock repository
func NewMockRepository[T any](getID func(T) uuid.UUID, setID func(*T, uuid.UUID)) *MockRepository[T] {
	return &MockRepository[T]{
		items: make(map[uuid.UUID]T),
		getID: getID,
		setID: setID,
	}
}

// Create adds an entity
func (r *MockRepository[T]) Create(ctx context.Context, entity *T) error {
	if r.ErrCreate != nil {
		return r.ErrCreate
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	id := uuid.New()
	r.setID(entity, id)
	r.items[id] = *entity
	return nil
}

// Update updates an entity
func (r *MockRepository[T]) Update(ctx context.Context, entity *T) error {
	if r.ErrUpdate != nil {
		return r.ErrUpdate
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	id := r.getID(*entity)
	if _, exists := r.items[id]; !exists {
		return errors.New("entity not found")
	}
	r.items[id] = *entity
	return nil
}

// Delete removes an entity
func (r *MockRepository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	if r.ErrDelete != nil {
		return r.ErrDelete
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.items, id)
	return nil
}

// FindByID finds an entity by ID
func (r *MockRepository[T]) FindByID(ctx context.Context, id uuid.UUID) (*T, error) {
	if r.ErrFind != nil {
		return nil, r.ErrFind
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	entity, exists := r.items[id]
	if !exists {
		return nil, nil
	}
	return &entity, nil
}

// FindAll returns all entities
func (r *MockRepository[T]) FindAll(ctx context.Context) ([]T, error) {
	if r.ErrList != nil {
		return nil, r.ErrList
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]T, 0, len(r.items))
	for _, item := range r.items {
		result = append(result, item)
	}
	return result, nil
}

// Count returns the number of entities
func (r *MockRepository[T]) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items)
}

// Reset clears all data and errors
func (r *MockRepository[T]) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = make(map[uuid.UUID]T)
	r.ErrCreate = nil
	r.ErrUpdate = nil
	r.ErrDelete = nil
	r.ErrFind = nil
	r.ErrList = nil
}

// ============================================
// Mock Cache
// ============================================

// MockCache is an in-memory cache for testing
type MockCache struct {
	mu    sync.RWMutex
	items map[string]mockCacheItem

	// Error injection
	ErrGet    error
	ErrSet    error
	ErrDelete error
}

type mockCacheItem struct {
	value     interface{}
	expiresAt time.Time
}

// NewMockCache creates a new mock cache
func NewMockCache() *MockCache {
	return &MockCache{
		items: make(map[string]mockCacheItem),
	}
}

// Get retrieves a value
func (c *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
	if c.ErrGet != nil {
		return c.ErrGet
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return errors.New("key not found")
	}
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		return errors.New("key expired")
	}

	// Simple copy for basic types
	switch d := dest.(type) {
	case *string:
		if v, ok := item.value.(string); ok {
			*d = v
		}
	case *int:
		if v, ok := item.value.(int); ok {
			*d = v
		}
	}
	return nil
}

// Set stores a value
func (c *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if c.ErrSet != nil {
		return c.ErrSet
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	item := mockCacheItem{value: value}
	if ttl > 0 {
		item.expiresAt = time.Now().Add(ttl)
	}
	c.items[key] = item
	return nil
}

// Delete removes a value
func (c *MockCache) Delete(ctx context.Context, key string) error {
	if c.ErrDelete != nil {
		return c.ErrDelete
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}

// Exists checks if key exists
func (c *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return false, nil
	}
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		return false, nil
	}
	return true, nil
}

// Reset clears all data
func (c *MockCache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]mockCacheItem)
	c.ErrGet = nil
	c.ErrSet = nil
	c.ErrDelete = nil
}

// ============================================
// Mock HTTP Client
// ============================================

// MockHTTPResponse represents a mocked response
type MockHTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
	Error      error
}

// MockHTTPClient is a mock HTTP client for testing
type MockHTTPClient struct {
	mu        sync.Mutex
	responses map[string]*MockHTTPResponse
	requests  []MockHTTPRequest
}

// MockHTTPRequest represents a captured request
type MockHTTPRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
}

// NewMockHTTPClient creates a new mock HTTP client
func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		responses: make(map[string]*MockHTTPResponse),
	}
}

// MockResponse sets up a mock response for a URL
func (c *MockHTTPClient) MockResponse(method, url string, resp *MockHTTPResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := method + " " + url
	c.responses[key] = resp
}

// GetRequests returns captured requests
func (c *MockHTTPClient) GetRequests() []MockHTTPRequest {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.requests
}

// Reset clears all mocks
func (c *MockHTTPClient) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.responses = make(map[string]*MockHTTPResponse)
	c.requests = nil
}
