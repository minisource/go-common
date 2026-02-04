package limiter

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// limiterEntry wraps a rate limiter with last access time
type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter manages rate limiters per IP address with automatic cleanup
type IPRateLimiter struct {
	ips       map[string]*limiterEntry
	mu        *sync.RWMutex
	r         rate.Limit
	b         int
	ttl       time.Duration
	stopClean chan struct{}
	stopped   bool
}

// Config holds configuration for IPRateLimiter
type Config struct {
	Rate       rate.Limit    // Requests per second
	Burst      int           // Maximum burst size
	TTL        time.Duration // Time to keep inactive IPs (default: 1 hour)
	CleanupInt time.Duration // Cleanup interval (default: 5 minutes)
}

// DefaultConfig returns default rate limiter configuration
func DefaultConfig() Config {
	return Config{
		Rate:       10,              // 10 requests per second
		Burst:      20,              // burst of 20
		TTL:        time.Hour,       // keep for 1 hour
		CleanupInt: 5 * time.Minute, // clean every 5 minutes
	}
}

// NewIPRateLimiter creates a new IP rate limiter with automatic cleanup
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return NewIPRateLimiterWithTTL(r, b, time.Hour, 5*time.Minute)
}

// NewIPRateLimiterWithTTL creates a new IP rate limiter with custom TTL
func NewIPRateLimiterWithTTL(r rate.Limit, b int, ttl, cleanupInterval time.Duration) *IPRateLimiter {
	if ttl == 0 {
		ttl = time.Hour
	}
	if cleanupInterval == 0 {
		cleanupInterval = 5 * time.Minute
	}

	i := &IPRateLimiter{
		ips:       make(map[string]*limiterEntry),
		mu:        &sync.RWMutex{},
		r:         r,
		b:         b,
		ttl:       ttl,
		stopClean: make(chan struct{}),
	}

	// Start background cleanup goroutine
	go i.cleanupLoop(cleanupInterval)

	return i
}

// NewIPRateLimiterFromConfig creates a new IP rate limiter from config
func NewIPRateLimiterFromConfig(cfg Config) *IPRateLimiter {
	return NewIPRateLimiterWithTTL(cfg.Rate, cfg.Burst, cfg.TTL, cfg.CleanupInt)
}

// cleanupLoop periodically removes expired entries
func (i *IPRateLimiter) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			i.cleanup()
		case <-i.stopClean:
			return
		}
	}
}

// cleanup removes entries that haven't been accessed within TTL
func (i *IPRateLimiter) cleanup() {
	i.mu.Lock()
	defer i.mu.Unlock()

	cutoff := time.Now().Add(-i.ttl)
	for ip, entry := range i.ips {
		if entry.lastSeen.Before(cutoff) {
			delete(i.ips, ip)
		}
	}
}

// Stop stops the cleanup goroutine
func (i *IPRateLimiter) Stop() {
	i.mu.Lock()
	defer i.mu.Unlock()

	if !i.stopped {
		close(i.stopClean)
		i.stopped = true
	}
}

// AddIP creates a new rate limiter and adds it to the ips map,
// using the IP address as the key
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)
	i.ips[ip] = &limiterEntry{
		limiter:  limiter,
		lastSeen: time.Now(),
	}

	return limiter
}

// GetLimiter returns the rate limiter for the provided IP address if it exists.
// Otherwise calls AddIP to add IP address to the map
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	entry, exists := i.ips[ip]

	if !exists {
		i.mu.Unlock()
		return i.AddIP(ip)
	}

	// Update last seen time
	entry.lastSeen = time.Now()
	i.mu.Unlock()

	return entry.limiter
}

// Len returns the number of tracked IPs
func (i *IPRateLimiter) Len() int {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return len(i.ips)
}

// Clear removes all entries
func (i *IPRateLimiter) Clear() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.ips = make(map[string]*limiterEntry)
}
