package retention

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// EligibilityFunc returns the count of records eligible for deletion according
// to the given snapshot. Services implement this to query their specific
// tables using cursor-based batching.
//
// For Age/Hybrid strategies: count records with created_at < snapshot.Cutoff.
// For Count/Hybrid strategies: count total records to determine how many
// exceed snapshot.KeepLatest.
type EligibilityFunc func(ctx context.Context, snapshot RunSnapshot) (int64, error)

// DeleteBatchFunc deletes one batch of eligible records and returns:
//   - deleted: number of records removed in this batch
//   - lastCreatedAt: the created_at of the last deleted record (for cursor)
//   - lastID: the primary key of the last deleted record (for cursor)
//   - hasMore: true if there may be more eligible records
//
// The cursor (lastCreatedAt, lastID) marks the boundary; the next batch must
// start AFTER this point using:
//
//	WHERE (created_at, id) > ($1, $2) AND created_at < $3
//	ORDER BY created_at ASC, id ASC LIMIT $4
type DeleteBatchFunc func(ctx context.Context, snapshot RunSnapshot, lastCreatedAt time.Time, lastID uuid.UUID) (deleted int64, newLastCreatedAt time.Time, newLastID uuid.UUID, hasMore bool, err error)

// RunResult captures the outcome of a cleanup run.
type RunResult struct {
	RunID        string
	Snapshot     RunSnapshot
	Result       Result
	ScannedCount int64
	DeletedCount int64
	BatchesRun   int
	StartedAt    time.Time
	EndedAt      time.Time
	Error        error
}

// BatchRunner orchestrates a safe, bounded, cursor-based deletion of eligible
// records. It implements:
//   - Deterministic ordering (created_at ASC, id ASC)
//   - Fixed cutoff frozen at run start
//   - Max batches per run
//   - Context cancellation
//   - Idempotent retry
//   - No OFFSET
type BatchRunner struct {
	snapshot      RunSnapshot
	eligibilityFn EligibilityFunc
	deleteFn      DeleteBatchFunc
}

// NewBatchRunner creates a BatchRunner.
func NewBatchRunner(snapshot RunSnapshot, eligibilityFn EligibilityFunc, deleteFn DeleteBatchFunc) *BatchRunner {
	return &BatchRunner{
		snapshot:      snapshot,
		eligibilityFn: eligibilityFn,
		deleteFn:      deleteFn,
	}
}

// Run executes the cleanup. If the policy is a dry-run it calls eligibilityFn
// and returns without deleting. Otherwise it deletes in batches.
func (r *BatchRunner) Run(ctx context.Context) RunResult {
	result := RunResult{
		RunID:     r.snapshot.RunID,
		Snapshot:  r.snapshot,
		StartedAt: time.Now().UTC(),
		Result:    ResultSuccess,
	}

	// Dry-run: preview only.
	if r.snapshot.DryRun {
		count, err := r.eligibilityFn(ctx, r.snapshot)
		if err != nil {
			result.Result = ResultFailed
			result.Error = err
			result.EndedAt = time.Now().UTC()
			return result
		}
		result.ScannedCount = count
		result.EndedAt = time.Now().UTC()
		result.Result = ResultSkippedDryRun
		return result
	}

	// Live run: delete in batches.
	var (
		lastCreatedAt time.Time
		lastID        = uuid.Nil
	)

	for batch := 0; batch < r.snapshot.MaxBatches; batch++ {
		select {
		case <-ctx.Done():
			result.Result = ResultPartial
			result.Error = ErrContextCancelled
			result.EndedAt = time.Now().UTC()
			return result
		default:
		}

		deleted, newCreatedAt, newID, hasMore, err := r.deleteFn(ctx, r.snapshot, lastCreatedAt, lastID)
		if err != nil {
			result.Result = ResultPartial
			// Detect context cancellation and wrap it for consistent error handling.
			if ctx.Err() != nil {
				result.Error = ErrContextCancelled
			} else {
				result.Error = err
			}
			result.EndedAt = time.Now().UTC()
			return result
		}

		result.DeletedCount += deleted
		result.BatchesRun++

		if !hasMore || deleted == 0 {
			break
		}

		lastCreatedAt = newCreatedAt
		lastID = newID
	}

	result.EndedAt = time.Now().UTC()

	// If we stopped because of batch limit, report partial.
	if result.BatchesRun >= r.snapshot.MaxBatches {
		result.Result = ResultPartial
		result.Error = ErrMaxBatchesReached
	}

	return result
}
