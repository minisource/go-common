package shutdown

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Hook represents a shutdown hook function
type Hook func(ctx context.Context) error

// Manager manages graceful shutdown
type Manager struct {
	mu      sync.RWMutex
	hooks   []namedHook
	timeout time.Duration
	signals []os.Signal
	done    chan struct{}
	started bool
}

type namedHook struct {
	name string
	fn   Hook
}

// NewManager creates a new shutdown manager
func NewManager(opts ...Option) *Manager {
	m := &Manager{
		timeout: 30 * time.Second,
		signals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
		done:    make(chan struct{}),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// Option configures the shutdown manager
type Option func(*Manager)

// WithTimeout sets the shutdown timeout
func WithTimeout(timeout time.Duration) Option {
	return func(m *Manager) {
		m.timeout = timeout
	}
}

// WithSignals sets the signals to listen for
func WithSignals(signals ...os.Signal) Option {
	return func(m *Manager) {
		m.signals = signals
	}
}

// AddHook adds a shutdown hook with a name
func (m *Manager) AddHook(name string, hook Hook) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.hooks = append(m.hooks, namedHook{name: name, fn: hook})
}

// AddFiberApp adds a Fiber app for graceful shutdown
func (m *Manager) AddFiberApp(name string, app *fiber.App) {
	m.AddHook(name, func(ctx context.Context) error {
		return app.ShutdownWithContext(ctx)
	})
}

// AddCloseFunc adds a close function as a hook
func (m *Manager) AddCloseFunc(name string, fn func() error) {
	m.AddHook(name, func(ctx context.Context) error {
		return fn()
	})
}

// Wait blocks until shutdown is triggered
func (m *Manager) Wait() {
	<-m.done
}

// Done returns a channel that's closed when shutdown completes
func (m *Manager) Done() <-chan struct{} {
	return m.done
}

// Start begins listening for shutdown signals
// Returns a function to trigger manual shutdown
func (m *Manager) Start() func() {
	m.mu.Lock()
	if m.started {
		m.mu.Unlock()
		return func() {}
	}
	m.started = true
	m.mu.Unlock()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, m.signals...)

	go func() {
		<-sigChan
		m.shutdown()
	}()

	return func() {
		m.shutdown()
	}
}

// shutdown executes all hooks in reverse order
func (m *Manager) shutdown() {
	m.mu.RLock()
	hooks := make([]namedHook, len(m.hooks))
	copy(hooks, m.hooks)
	m.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	// Execute hooks in reverse order (LIFO)
	for i := len(hooks) - 1; i >= 0; i-- {
		hook := hooks[i]
		if err := hook.fn(ctx); err != nil {
			// Log error but continue with other hooks
			continue
		}
	}

	close(m.done)
}

// ============================================
// Convenience Functions
// ============================================

// DefaultManager is a global shutdown manager
var defaultManager = NewManager()

// Add adds a hook to the default manager
func Add(name string, hook Hook) {
	defaultManager.AddHook(name, hook)
}

// AddCloseFunc adds a close function to the default manager
func AddClose(name string, fn func() error) {
	defaultManager.AddCloseFunc(name, fn)
}

// AddFiber adds a Fiber app to the default manager
func AddFiber(name string, app *fiber.App) {
	defaultManager.AddFiberApp(name, app)
}

// Start starts the default manager
func Start() func() {
	return defaultManager.Start()
}

// Wait waits for shutdown on the default manager
func Wait() {
	defaultManager.Wait()
}

// ============================================
// Health-Aware Shutdown
// ============================================

// HealthAwareManager extends Manager with health awareness
type HealthAwareManager struct {
	*Manager
	healthCheckInterval time.Duration
	preShutdownDelay    time.Duration
	isHealthy           bool
	mu                  sync.RWMutex
}

// NewHealthAwareManager creates a health-aware shutdown manager
func NewHealthAwareManager(opts ...Option) *HealthAwareManager {
	return &HealthAwareManager{
		Manager:             NewManager(opts...),
		healthCheckInterval: 5 * time.Second,
		preShutdownDelay:    5 * time.Second,
		isHealthy:           true,
	}
}

// SetHealthy sets the health status
func (m *HealthAwareManager) SetHealthy(healthy bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isHealthy = healthy
}

// IsHealthy returns the current health status
func (m *HealthAwareManager) IsHealthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isHealthy
}

// WithPreShutdownDelay sets delay before starting shutdown hooks
// This allows load balancers to stop sending traffic
func (m *HealthAwareManager) WithPreShutdownDelay(d time.Duration) *HealthAwareManager {
	m.preShutdownDelay = d
	return m
}

// GracefulShutdown performs health-aware graceful shutdown
func (m *HealthAwareManager) GracefulShutdown() {
	// Mark as unhealthy first
	m.SetHealthy(false)

	// Wait for load balancers to stop traffic
	time.Sleep(m.preShutdownDelay)

	// Then proceed with normal shutdown
	m.shutdown()
}
