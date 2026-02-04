package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements Cache interface using Redis
type RedisCache struct {
	client  *redis.Client
	options Options
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(client *redis.Client, opts ...Options) *RedisCache {
	options := DefaultOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	return &RedisCache{
		client:  client,
		options: options,
	}
}

// buildKey builds a key with prefix
func (c *RedisCache) buildKey(key string) string {
	if c.options.KeyPrefix != "" {
		return c.options.KeyPrefix + ":" + key
	}
	return key
}

// Get retrieves a value by key
func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	result, err := c.client.Get(ctx, c.buildKey(key)).Bytes()
	if err == redis.Nil {
		return nil, ErrKeyNotFound
	}
	return result, err
}

// GetObject retrieves and unmarshals a value
func (c *RedisCache) GetObject(ctx context.Context, key string, dest interface{}) error {
	data, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	return c.options.Serializer.Unmarshal(data, dest)
}

// Set stores a value with TTL
func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.options.DefaultTTL
	}
	return c.client.Set(ctx, c.buildKey(key), value, ttl).Err()
}

// SetObject marshals and stores a value
func (c *RedisCache) SetObject(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := c.options.Serializer.Marshal(value)
	if err != nil {
		return err
	}
	return c.Set(ctx, key, data, ttl)
}

// Delete removes a key
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.buildKey(key)).Err()
}

// Exists checks if key exists
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, c.buildKey(key)).Result()
	return n > 0, err
}

// TTL returns remaining TTL for key
func (c *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.client.TTL(ctx, c.buildKey(key)).Result()
	if err != nil {
		return 0, err
	}
	if ttl < 0 {
		return 0, ErrKeyNotFound
	}
	return ttl, nil
}

// Increment increments a numeric value
func (c *RedisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return c.client.IncrBy(ctx, c.buildKey(key), delta).Result()
}

// Decrement decrements a numeric value
func (c *RedisCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return c.client.DecrBy(ctx, c.buildKey(key), delta).Result()
}

// SetNX sets value only if not exists
func (c *RedisCache) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	if ttl == 0 {
		ttl = c.options.DefaultTTL
	}
	return c.client.SetNX(ctx, c.buildKey(key), value, ttl).Result()
}

// GetSet sets new value and returns old value
func (c *RedisCache) GetSet(ctx context.Context, key string, value []byte) ([]byte, error) {
	result, err := c.client.GetSet(ctx, c.buildKey(key), value).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return result, err
}

// Keys returns keys matching pattern
func (c *RedisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	fullPattern := c.buildKey(pattern)
	return c.client.Keys(ctx, fullPattern).Result()
}

// DeleteMany deletes multiple keys
func (c *RedisCache) DeleteMany(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = c.buildKey(key)
	}
	return c.client.Del(ctx, fullKeys...).Err()
}

// Ping checks connection
func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close closes the connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// ============================================
// Hash Operations
// ============================================

// HSet sets a hash field
func (c *RedisCache) HSet(ctx context.Context, key, field string, value []byte) error {
	return c.client.HSet(ctx, c.buildKey(key), field, value).Err()
}

// HGet gets a hash field
func (c *RedisCache) HGet(ctx context.Context, key, field string) ([]byte, error) {
	result, err := c.client.HGet(ctx, c.buildKey(key), field).Bytes()
	if err == redis.Nil {
		return nil, ErrKeyNotFound
	}
	return result, err
}

// HGetAll gets all hash fields
func (c *RedisCache) HGetAll(ctx context.Context, key string) (map[string][]byte, error) {
	result, err := c.client.HGetAll(ctx, c.buildKey(key)).Result()
	if err != nil {
		return nil, err
	}

	m := make(map[string][]byte)
	for k, v := range result {
		m[k] = []byte(v)
	}
	return m, nil
}

// HDel deletes hash fields
func (c *RedisCache) HDel(ctx context.Context, key string, fields ...string) error {
	return c.client.HDel(ctx, c.buildKey(key), fields...).Err()
}

// HExists checks if hash field exists
func (c *RedisCache) HExists(ctx context.Context, key, field string) (bool, error) {
	return c.client.HExists(ctx, c.buildKey(key), field).Result()
}

// ============================================
// List Operations
// ============================================

// LPush prepends values to list
func (c *RedisCache) LPush(ctx context.Context, key string, values ...[]byte) error {
	args := make([]interface{}, len(values))
	for i, v := range values {
		args[i] = v
	}
	return c.client.LPush(ctx, c.buildKey(key), args...).Err()
}

// RPush appends values to list
func (c *RedisCache) RPush(ctx context.Context, key string, values ...[]byte) error {
	args := make([]interface{}, len(values))
	for i, v := range values {
		args[i] = v
	}
	return c.client.RPush(ctx, c.buildKey(key), args...).Err()
}

// LPop removes and returns first element
func (c *RedisCache) LPop(ctx context.Context, key string) ([]byte, error) {
	result, err := c.client.LPop(ctx, c.buildKey(key)).Bytes()
	if err == redis.Nil {
		return nil, ErrKeyNotFound
	}
	return result, err
}

// RPop removes and returns last element
func (c *RedisCache) RPop(ctx context.Context, key string) ([]byte, error) {
	result, err := c.client.RPop(ctx, c.buildKey(key)).Bytes()
	if err == redis.Nil {
		return nil, ErrKeyNotFound
	}
	return result, err
}

// LRange returns range of elements
func (c *RedisCache) LRange(ctx context.Context, key string, start, stop int64) ([][]byte, error) {
	result, err := c.client.LRange(ctx, c.buildKey(key), start, stop).Result()
	if err != nil {
		return nil, err
	}

	bytes := make([][]byte, len(result))
	for i, v := range result {
		bytes[i] = []byte(v)
	}
	return bytes, nil
}

// LLen returns list length
func (c *RedisCache) LLen(ctx context.Context, key string) (int64, error) {
	return c.client.LLen(ctx, c.buildKey(key)).Result()
}

// ============================================
// Set Operations
// ============================================

// SAdd adds members to set
func (c *RedisCache) SAdd(ctx context.Context, key string, members ...[]byte) error {
	args := make([]interface{}, len(members))
	for i, m := range members {
		args[i] = m
	}
	return c.client.SAdd(ctx, c.buildKey(key), args...).Err()
}

// SRem removes members from set
func (c *RedisCache) SRem(ctx context.Context, key string, members ...[]byte) error {
	args := make([]interface{}, len(members))
	for i, m := range members {
		args[i] = m
	}
	return c.client.SRem(ctx, c.buildKey(key), args...).Err()
}

// SMembers returns all set members
func (c *RedisCache) SMembers(ctx context.Context, key string) ([][]byte, error) {
	result, err := c.client.SMembers(ctx, c.buildKey(key)).Result()
	if err != nil {
		return nil, err
	}

	bytes := make([][]byte, len(result))
	for i, v := range result {
		bytes[i] = []byte(v)
	}
	return bytes, nil
}

// SIsMember checks if member exists in set
func (c *RedisCache) SIsMember(ctx context.Context, key string, member []byte) (bool, error) {
	return c.client.SIsMember(ctx, c.buildKey(key), member).Result()
}

// SCard returns set cardinality
func (c *RedisCache) SCard(ctx context.Context, key string) (int64, error) {
	return c.client.SCard(ctx, c.buildKey(key)).Result()
}
