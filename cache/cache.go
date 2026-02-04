package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExpired  = errors.New("key expired")
)

// Cache defines the cache interface
type Cache interface {
	// Get retrieves a value by key
	Get(ctx context.Context, key string) ([]byte, error)

	// GetObject retrieves and unmarshals a value
	GetObject(ctx context.Context, key string, dest interface{}) error

	// Set stores a value with optional TTL
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// SetObject marshals and stores a value
	SetObject(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a key
	Delete(ctx context.Context, key string) error

	// Exists checks if key exists
	Exists(ctx context.Context, key string) (bool, error)

	// TTL returns remaining TTL for key
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Increment increments a numeric value
	Increment(ctx context.Context, key string, delta int64) (int64, error)

	// Decrement decrements a numeric value
	Decrement(ctx context.Context, key string, delta int64) (int64, error)

	// SetNX sets value only if not exists (returns true if set)
	SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)

	// GetSet sets new value and returns old value
	GetSet(ctx context.Context, key string, value []byte) ([]byte, error)

	// Keys returns keys matching pattern
	Keys(ctx context.Context, pattern string) ([]string, error)

	// DeleteMany deletes multiple keys
	DeleteMany(ctx context.Context, keys ...string) error

	// Ping checks connection
	Ping(ctx context.Context) error

	// Close closes the connection
	Close() error
}

// HashCache defines hash operations
type HashCache interface {
	// HSet sets a hash field
	HSet(ctx context.Context, key, field string, value []byte) error

	// HGet gets a hash field
	HGet(ctx context.Context, key, field string) ([]byte, error)

	// HGetAll gets all hash fields
	HGetAll(ctx context.Context, key string) (map[string][]byte, error)

	// HDel deletes hash fields
	HDel(ctx context.Context, key string, fields ...string) error

	// HExists checks if hash field exists
	HExists(ctx context.Context, key, field string) (bool, error)
}

// ListCache defines list operations
type ListCache interface {
	// LPush prepends values to list
	LPush(ctx context.Context, key string, values ...[]byte) error

	// RPush appends values to list
	RPush(ctx context.Context, key string, values ...[]byte) error

	// LPop removes and returns first element
	LPop(ctx context.Context, key string) ([]byte, error)

	// RPop removes and returns last element
	RPop(ctx context.Context, key string) ([]byte, error)

	// LRange returns range of elements
	LRange(ctx context.Context, key string, start, stop int64) ([][]byte, error)

	// LLen returns list length
	LLen(ctx context.Context, key string) (int64, error)
}

// SetCache defines set operations
type SetCache interface {
	// SAdd adds members to set
	SAdd(ctx context.Context, key string, members ...[]byte) error

	// SRem removes members from set
	SRem(ctx context.Context, key string, members ...[]byte) error

	// SMembers returns all set members
	SMembers(ctx context.Context, key string) ([][]byte, error)

	// SIsMember checks if member exists in set
	SIsMember(ctx context.Context, key string, member []byte) (bool, error)

	// SCard returns set cardinality
	SCard(ctx context.Context, key string) (int64, error)
}

// ============================================
// Cache Options
// ============================================

// Options configures cache behavior
type Options struct {
	// DefaultTTL is the default TTL for keys without explicit TTL
	DefaultTTL time.Duration

	// KeyPrefix is prepended to all keys
	KeyPrefix string

	// Serializer customizes value serialization
	Serializer Serializer
}

// DefaultOptions returns default cache options
func DefaultOptions() Options {
	return Options{
		DefaultTTL: 24 * time.Hour,
		KeyPrefix:  "",
		Serializer: &JSONSerializer{},
	}
}

// ============================================
// Serializers
// ============================================

// Serializer defines serialization interface
type Serializer interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

// JSONSerializer uses JSON for serialization
type JSONSerializer struct{}

// Marshal serializes to JSON
func (s *JSONSerializer) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal deserializes from JSON
func (s *JSONSerializer) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// ============================================
// Cache Key Builder
// ============================================

// KeyBuilder helps build cache keys
type KeyBuilder struct {
	prefix string
}

// NewKeyBuilder creates a new key builder
func NewKeyBuilder(prefix string) *KeyBuilder {
	return &KeyBuilder{prefix: prefix}
}

// Key builds a key with parts
func (b *KeyBuilder) Key(parts ...string) string {
	key := b.prefix
	for _, part := range parts {
		if key != "" {
			key += ":"
		}
		key += part
	}
	return key
}

// UserKey builds a user-scoped key
func (b *KeyBuilder) UserKey(userID, key string) string {
	return b.Key("user", userID, key)
}

// TenantKey builds a tenant-scoped key
func (b *KeyBuilder) TenantKey(tenantID, key string) string {
	return b.Key("tenant", tenantID, key)
}

// SessionKey builds a session key
func (b *KeyBuilder) SessionKey(sessionID string) string {
	return b.Key("session", sessionID)
}

// ============================================
// GetOrSet Helper
// ============================================

// GetOrSet gets value from cache or computes and stores it
func GetOrSet[T any](ctx context.Context, cache Cache, key string, ttl time.Duration, compute func() (T, error)) (T, error) {
	var result T

	// Try to get from cache
	err := cache.GetObject(ctx, key, &result)
	if err == nil {
		return result, nil
	}

	// Compute the value
	result, err = compute()
	if err != nil {
		return result, err
	}

	// Store in cache (ignore error)
	_ = cache.SetObject(ctx, key, result, ttl)

	return result, nil
}
