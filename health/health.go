package health

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// CheckResult represents the result of a health check
type CheckResult struct {
	Name      string        `json:"name"`
	Status    Status        `json:"status"`
	Message   string        `json:"message,omitempty"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
}

// Checker interface for health checks
type Checker interface {
	Name() string
	Check(ctx context.Context) error
}

// HealthService manages health checks
type HealthService struct {
	checkers []Checker
	mu       sync.RWMutex
	timeout  time.Duration
}

// Config for health service
type Config struct {
	Timeout time.Duration
}

// DefaultConfig returns default health config
func DefaultConfig() Config {
	return Config{
		Timeout: 5 * time.Second,
	}
}

// NewHealthService creates a new health service
func NewHealthService(cfg Config) *HealthService {
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	return &HealthService{
		checkers: make([]Checker, 0),
		timeout:  cfg.Timeout,
	}
}

// RegisterChecker adds a checker to the health service
func (h *HealthService) RegisterChecker(checker Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers = append(h.checkers, checker)
}

// CheckLiveness performs a basic liveness check
func (h *HealthService) CheckLiveness() map[string]interface{} {
	return map[string]interface{}{
		"status":    StatusHealthy,
		"timestamp": time.Now(),
	}
}

// CheckReadiness performs all registered health checks
func (h *HealthService) CheckReadiness(ctx context.Context) (map[string]interface{}, bool) {
	h.mu.RLock()
	checkers := make([]Checker, len(h.checkers))
	copy(checkers, h.checkers)
	h.mu.RUnlock()

	if len(checkers) == 0 {
		return map[string]interface{}{
			"status":    StatusHealthy,
			"timestamp": time.Now(),
			"checks":    []CheckResult{},
		}, true
	}

	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	results := make([]CheckResult, len(checkers))
	var wg sync.WaitGroup
	allHealthy := true

	for i, checker := range checkers {
		wg.Add(1)
		go func(idx int, c Checker) {
			defer wg.Done()

			start := time.Now()
			err := c.Check(ctx)
			duration := time.Since(start)

			result := CheckResult{
				Name:      c.Name(),
				Duration:  duration,
				Timestamp: time.Now(),
			}

			if err != nil {
				result.Status = StatusUnhealthy
				result.Message = err.Error()
				allHealthy = false
			} else {
				result.Status = StatusHealthy
			}

			results[idx] = result
		}(i, checker)
	}

	wg.Wait()

	status := StatusHealthy
	if !allHealthy {
		status = StatusUnhealthy
	}

	return map[string]interface{}{
		"status":    status,
		"timestamp": time.Now(),
		"checks":    results,
	}, allHealthy
}

// PostgresChecker checks PostgreSQL connectivity
type PostgresChecker struct {
	name   string
	pinger Pinger
}

// Pinger interface for database ping
type Pinger interface {
	PingContext(ctx context.Context) error
}

// NewPostgresChecker creates a new postgres health checker
func NewPostgresChecker(name string, pinger Pinger) *PostgresChecker {
	return &PostgresChecker{
		name:   name,
		pinger: pinger,
	}
}

func (c *PostgresChecker) Name() string {
	return c.name
}

func (c *PostgresChecker) Check(ctx context.Context) error {
	if c.pinger == nil {
		return fmt.Errorf("database connection not initialized")
	}
	return c.pinger.PingContext(ctx)
}

// RedisChecker checks Redis connectivity
type RedisChecker struct {
	name   string
	pinger RedisPinger
}

// RedisPinger interface for redis ping
type RedisPinger interface {
	Ping(ctx context.Context) error
}

// NewRedisChecker creates a new redis health checker
func NewRedisChecker(name string, pinger RedisPinger) *RedisChecker {
	return &RedisChecker{
		name:   name,
		pinger: pinger,
	}
}

func (c *RedisChecker) Name() string {
	return c.name
}

func (c *RedisChecker) Check(ctx context.Context) error {
	if c.pinger == nil {
		return fmt.Errorf("redis connection not initialized")
	}
	return c.pinger.Ping(ctx)
}

// CustomChecker wraps a function as a health checker
type CustomChecker struct {
	name      string
	checkFunc func(ctx context.Context) error
}

// NewCustomChecker creates a custom health checker
func NewCustomChecker(name string, checkFunc func(ctx context.Context) error) *CustomChecker {
	return &CustomChecker{
		name:      name,
		checkFunc: checkFunc,
	}
}

func (c *CustomChecker) Name() string {
	return c.name
}

func (c *CustomChecker) Check(ctx context.Context) error {
	if c.checkFunc == nil {
		return nil
	}
	return c.checkFunc(ctx)
}
