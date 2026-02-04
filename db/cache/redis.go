package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v7"
)

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host               string        `env:"REDIS_HOST"`
	Port               string        `env:"REDIS_PORT"`
	Password           string        `env:"REDIS_PASSWORD"`
	Db                 string        `env:"REDIS_DB"`
	DialTimeout        time.Duration `env:"REDIS_DIALTIMEOUT"`
	ReadTimeout        time.Duration `env:"REDIS_READTIMEOUT"`
	WriteTimeout       time.Duration `env:"REDIS_WRITETIMEOUT"`
	IdleCheckFrequency time.Duration `env:"REDIS_IDLECHECKFREQUENCY"`
	PoolSize           int           `env:"REDIS_POOL_SIZE"`
	PoolTimeout        time.Duration `env:"REDIS_POOL_TIMEOUT"`
}

// RedisClient wraps the redis.Client
type RedisClient struct {
	client *redis.Client
	cfg    *RedisConfig
}

// NewRedisClient creates and returns a new RedisClient instance
func NewRedisClient(cfg *RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:               fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:           cfg.Password,
		DB:                 0,
		DialTimeout:        cfg.DialTimeout * time.Second,
		ReadTimeout:        cfg.ReadTimeout * time.Second,
		WriteTimeout:       cfg.WriteTimeout * time.Second,
		PoolSize:           cfg.PoolSize,
		PoolTimeout:        cfg.PoolTimeout,
		IdleTimeout:        500 * time.Millisecond,
		IdleCheckFrequency: cfg.IdleCheckFrequency * time.Millisecond,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisClient{client: client, cfg: cfg}, nil
}

// Client returns the underlying redis.Client
func (r *RedisClient) Client() *redis.Client {
	return r.client
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Ping checks Redis connectivity
func (r *RedisClient) Ping(ctx context.Context) error {
	_, err := r.client.Ping().Result()
	return err
}

// Deprecated: Use NewRedisClient instead
// The following functions are kept for backward compatibility

var redisClient *redis.Client

// InitRedis initializes the Redis connection (deprecated: use NewRedisClient)
func InitRedis(cfg *RedisConfig) error {
	redisClient = redis.NewClient(&redis.Options{
		Addr:               fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:           cfg.Password,
		DB:                 0,
		DialTimeout:        cfg.DialTimeout * time.Second,
		ReadTimeout:        cfg.ReadTimeout * time.Second,
		WriteTimeout:       cfg.WriteTimeout * time.Second,
		PoolSize:           cfg.PoolSize,
		PoolTimeout:        cfg.PoolTimeout,
		IdleTimeout:        500 * time.Millisecond,
		IdleCheckFrequency: cfg.IdleCheckFrequency * time.Millisecond,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		return err
	}
	return nil
}

// GetRedis returns the Redis client (deprecated: use NewRedisClient)
func GetRedis() *redis.Client {
	return redisClient
}

// CloseRedis closes the Redis connection (deprecated: use RedisClient.Close)
func CloseRedis() {
	redisClient.Close()
}

func Set[T any](c *redis.Client, key string, value T, duration time.Duration) error {
	v, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Set(key, v, duration).Err()
}

func Get[T any](c *redis.Client, key string) (T, error) {
	var dest T = *new(T)
	v, err := c.Get(key).Result()
	if err != nil {
		return dest, err
	}
	err = json.Unmarshal([]byte(v), &dest)
	if err != nil {
		return dest, err
	}
	return dest, nil
}
