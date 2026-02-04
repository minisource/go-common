package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClientV9 wraps the go-redis v9 client
type RedisClientV9 struct {
	client *redis.Client
	cfg    *RedisConfigV9
}

// RedisConfigV9 holds configuration for Redis v9 client
type RedisConfigV9 struct {
	Host               string
	Port               string
	Password           string
	DB                 int
	DialTimeout        time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	PoolSize           int
	MinIdleConns       int
	MaxConnAge         time.Duration
	PoolTimeout        time.Duration
	IdleTimeout        time.Duration
	IdleCheckFrequency time.Duration
}

// DefaultRedisConfigV9 returns default Redis configuration
func DefaultRedisConfigV9() *RedisConfigV9 {
	return &RedisConfigV9{
		Host:               "localhost",
		Port:               "6379",
		Password:           "",
		DB:                 0,
		DialTimeout:        5 * time.Second,
		ReadTimeout:        3 * time.Second,
		WriteTimeout:       3 * time.Second,
		PoolSize:           10,
		MinIdleConns:       2,
		MaxConnAge:         0,
		PoolTimeout:        4 * time.Second,
		IdleTimeout:        5 * time.Minute,
		IdleCheckFrequency: time.Minute,
	}
}

// NewRedisClientV9 creates a new Redis v9 client
func NewRedisClientV9(ctx context.Context, cfg *RedisConfigV9) (*RedisClientV9, error) {
	if cfg == nil {
		cfg = DefaultRedisConfigV9()
	}

	client := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:        cfg.Password,
		DB:              cfg.DB,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		ConnMaxLifetime: cfg.MaxConnAge,
		PoolTimeout:     cfg.PoolTimeout,
		ConnMaxIdleTime: cfg.IdleTimeout,
	})

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisClientV9{
		client: client,
		cfg:    cfg,
	}, nil
}

// Client returns the underlying redis client
func (r *RedisClientV9) Client() *redis.Client {
	return r.client
}

// Close closes the Redis connection
func (r *RedisClientV9) Close() error {
	return r.client.Close()
}

// Ping checks Redis connectivity
func (r *RedisClientV9) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Set stores a value with expiration
func (r *RedisClientV9) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return r.client.Set(ctx, key, data, expiration).Err()
}

// Get retrieves a value
func (r *RedisClientV9) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// GetString retrieves a string value
func (r *RedisClientV9) GetString(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Delete removes a key
func (r *RedisClientV9) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists checks if keys exist
func (r *RedisClientV9) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

// Expire sets expiration on a key
func (r *RedisClientV9) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// TTL gets remaining time to live
func (r *RedisClientV9) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// Incr increments a key
func (r *RedisClientV9) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// Decr decrements a key
func (r *RedisClientV9) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}

// HSet sets hash field
func (r *RedisClientV9) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

// HGet gets hash field
func (r *RedisClientV9) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HGetAll gets all hash fields
func (r *RedisClientV9) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HDel deletes hash fields
func (r *RedisClientV9) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// SAdd adds members to a set
func (r *RedisClientV9) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

// SMembers gets all set members
func (r *RedisClientV9) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

// SRem removes members from a set
func (r *RedisClientV9) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SRem(ctx, key, members...).Err()
}

// SIsMember checks if member exists in set
func (r *RedisClientV9) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.client.SIsMember(ctx, key, member).Result()
}

// LPush prepends to a list
func (r *RedisClientV9) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

// RPush appends to a list
func (r *RedisClientV9) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.RPush(ctx, key, values...).Err()
}

// LRange gets list range
func (r *RedisClientV9) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.LRange(ctx, key, start, stop).Result()
}

// LLen gets list length
func (r *RedisClientV9) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, key).Result()
}

// Publish publishes a message to a channel
func (r *RedisClientV9) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to channels
func (r *RedisClientV9) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// SetNX sets value if not exists
func (r *RedisClientV9) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}
	return r.client.SetNX(ctx, key, data, expiration).Result()
}

// Keys returns keys matching pattern
func (r *RedisClientV9) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}

// FlushDB flushes the current database
func (r *RedisClientV9) FlushDB(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// Pipeline creates a pipeline
func (r *RedisClientV9) Pipeline() redis.Pipeliner {
	return r.client.Pipeline()
}

// TxPipeline creates a transaction pipeline
func (r *RedisClientV9) TxPipeline() redis.Pipeliner {
	return r.client.TxPipeline()
}

// IsNil checks if error is redis.Nil
func IsNil(err error) bool {
	return err == redis.Nil
}
