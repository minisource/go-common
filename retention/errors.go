package retention

import "errors"

// Sentinel errors returned by the retention package.
var (
	// ErrDisabled is returned when a cleanup is attempted against a disabled policy.
	ErrDisabled = errors.New("retention policy is disabled")

	// ErrDryRun is returned after a dry-run completes successfully without
	// deleting any records. Callers can use errors.Is to distinguish preview
	// from real execution.
	ErrDryRun = errors.New("dry-run completed: no records deleted")

	// ErrLockHeld is returned when the distributed lock is already held by
	// another instance.
	ErrLockHeld = errors.New("cleanup lock is held by another instance")

	// ErrMaxBatchesReached is returned when the runner hits the per-run batch
	// limit. It is not a failure — the run can be resumed on the next tick.
	ErrMaxBatchesReached = errors.New("max batches reached: run will resume next cycle")

	// ErrCategoryProtected is returned when a policy targets a protected
	// category that must never be cleaned.
	ErrCategoryProtected = errors.New("category is protected and cannot be cleaned")

	// ErrInvalidPolicy is returned when policy validation fails.
	ErrInvalidPolicy = errors.New("invalid retention policy")

	// ErrContextCancelled is returned when the context is cancelled mid-run.
	ErrContextCancelled = errors.New("cleanup context cancelled")

	// ErrPolicyNotFound is returned when a referenced policy does not exist.
	ErrPolicyNotFound = errors.New("retention policy not found")
)
