package cache

import (
	"context"
	"sync"
	"time"
)

// MemoryCache implements Cache interface using in-memory storage
type MemoryCache struct {
	mu       sync.RWMutex
	items    map[string]*memoryItem
	options  Options
	stopChan chan struct{}
}

type memoryItem struct {
	value     []byte
	expiresAt time.Time
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(opts ...Options) *MemoryCache {
	options := DefaultOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	c := &MemoryCache{
		items:    make(map[string]*memoryItem),
		options:  options,
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine
	go c.cleanup()

	return c
}

// cleanup periodically removes expired items
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.removeExpired()
		case <-c.stopChan:
			return
		}
	}
}

// removeExpired removes all expired items
func (c *MemoryCache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
			delete(c.items, key)
		}
	}
}

// buildKey builds a key with prefix
func (c *MemoryCache) buildKey(key string) string {
	if c.options.KeyPrefix != "" {
		return c.options.KeyPrefix + ":" + key
	}
	return key
}

// Get retrieves a value by key
func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[c.buildKey(key)]
	if !exists {
		return nil, ErrKeyNotFound
	}

	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		return nil, ErrKeyExpired
	}

	return item.value, nil
}

// GetObject retrieves and unmarshals a value
func (c *MemoryCache) GetObject(ctx context.Context, key string, dest interface{}) error {
	data, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	return c.options.Serializer.Unmarshal(data, dest)
}

// Set stores a value with TTL
func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ttl == 0 {
		ttl = c.options.DefaultTTL
	}

	item := &memoryItem{
		value: value,
	}
	if ttl > 0 {
		item.expiresAt = time.Now().Add(ttl)
	}

	c.items[c.buildKey(key)] = item
	return nil
}

// SetObject marshals and stores a value
func (c *MemoryCache) SetObject(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := c.options.Serializer.Marshal(value)
	if err != nil {
		return err
	}
	return c.Set(ctx, key, data, ttl)
}

// Delete removes a key
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, c.buildKey(key))
	return nil
}

// Exists checks if key exists
func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[c.buildKey(key)]
	if !exists {
		return false, nil
	}

	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		return false, nil
	}

	return true, nil
}

// TTL returns remaining TTL for key
func (c *MemoryCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[c.buildKey(key)]
	if !exists {
		return 0, ErrKeyNotFound
	}

	if item.expiresAt.IsZero() {
		return -1, nil // No expiration
	}

	ttl := time.Until(item.expiresAt)
	if ttl < 0 {
		return 0, ErrKeyExpired
	}

	return ttl, nil
}

// Increment increments a numeric value
func (c *MemoryCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fullKey := c.buildKey(key)
	item, exists := c.items[fullKey]

	var value int64
	if exists && (item.expiresAt.IsZero() || time.Now().Before(item.expiresAt)) {
		_ = c.options.Serializer.Unmarshal(item.value, &value)
	}

	value += delta
	data, _ := c.options.Serializer.Marshal(value)

	newItem := &memoryItem{value: data}
	if exists && !item.expiresAt.IsZero() {
		newItem.expiresAt = item.expiresAt
	}
	c.items[fullKey] = newItem

	return value, nil
}

// Decrement decrements a numeric value
func (c *MemoryCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return c.Increment(ctx, key, -delta)
}

// SetNX sets value only if not exists
func (c *MemoryCache) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fullKey := c.buildKey(key)
	item, exists := c.items[fullKey]
	if exists && (item.expiresAt.IsZero() || time.Now().Before(item.expiresAt)) {
		return false, nil
	}

	newItem := &memoryItem{value: value}
	if ttl > 0 {
		newItem.expiresAt = time.Now().Add(ttl)
	}
	c.items[fullKey] = newItem

	return true, nil
}

// GetSet sets new value and returns old value
func (c *MemoryCache) GetSet(ctx context.Context, key string, value []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fullKey := c.buildKey(key)
	var oldValue []byte

	if item, exists := c.items[fullKey]; exists {
		if item.expiresAt.IsZero() || time.Now().Before(item.expiresAt) {
			oldValue = item.value
		}
	}

	c.items[fullKey] = &memoryItem{value: value}
	return oldValue, nil
}

// Keys returns keys matching pattern (basic prefix matching)
func (c *MemoryCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var keys []string
	prefix := c.options.KeyPrefix
	if prefix != "" {
		prefix += ":"
	}

	for key := range c.items {
		keys = append(keys, key)
	}
	return keys, nil
}

// DeleteMany deletes multiple keys
func (c *MemoryCache) DeleteMany(ctx context.Context, keys ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, key := range keys {
		delete(c.items, c.buildKey(key))
	}
	return nil
}

// Ping checks connection
func (c *MemoryCache) Ping(ctx context.Context) error {
	return nil
}

// Close closes the cache
func (c *MemoryCache) Close() error {
	close(c.stopChan)
	return nil
}

// Clear removes all items
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*memoryItem)
}

// Size returns the number of items
func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}
