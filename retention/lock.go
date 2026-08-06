package retention

import (
	"context"
	"time"
)

// DistributedLock provides a service-level mutual exclusion mechanism so that
// only one replica executes a cleanup run for a given (service, category) pair
// at any time.
//
// Implementations must:
//   - Have a bounded lease (TTL) so that a crashed process does not hold the
//     lock forever.
//   - Support release on normal completion.
//   - Distinguish "already held" from a genuine acquisition error.
type DistributedLock interface {
	// Acquire attempts to obtain the lock for the given key. Returns a
	// non-nil LockGuard on success. The caller MUST call Release on the
	// guard when the protected work is done.
	Acquire(ctx context.Context, key string, ttl time.Duration) (LockGuard, error)

	// IsHeld reports whether the lock for key is currently held (best-effort).
	IsHeld(ctx context.Context, key string) (bool, error)
}

// LockGuard represents a held distributed lock. The caller is responsible for
// calling Release exactly once.
type LockGuard interface {
	// Release releases the lock. Safe to call multiple times (no-op after
	// first release). Must not return an error if the lock was already
	// released or expired.
	Release(ctx context.Context) error

	// Key returns the lock key.
	Key() string
}
