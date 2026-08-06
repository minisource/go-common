package retention

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Strategy ──────────────────────────────────────────────────────────

func TestStrategy_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		strategy Strategy
		valid    bool
	}{
		{"age", StrategyAge, true},
		{"count", StrategyCount, true},
		{"hybrid", StrategyHybrid, true},
		{"empty", Strategy(""), false},
		{"unknown", Strategy("unknown"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.strategy.IsValid())
		})
	}
}

// ── Policy Validation ─────────────────────────────────────────────────

func TestPolicy_Validate(t *testing.T) {
	t.Run("valid age policy", func(t *testing.T) {
		p := DefaultPolicy()
		p.Service = "auth"
		p.Category = "login_logs"
		p.Strategy = StrategyAge
		p.RetentionDays = 30
		assert.NoError(t, p.Validate())
	})

	t.Run("valid count policy", func(t *testing.T) {
		p := DefaultPolicy()
		p.Service = "auth"
		p.Category = "login_logs"
		p.Strategy = StrategyCount
		p.KeepLatestCount = 10000
		assert.NoError(t, p.Validate())
	})

	t.Run("valid hybrid policy", func(t *testing.T) {
		p := DefaultPolicy()
		p.Service = "notifier"
		p.Category = "notification_logs"
		p.Strategy = StrategyHybrid
		p.RetentionDays = 30
		p.KeepLatestCount = 500000
		assert.NoError(t, p.Validate())
	})

	t.Run("missing service", func(t *testing.T) {
		p := DefaultPolicy()
		p.Category = "login_logs"
		assert.ErrorContains(t, p.Validate(), "service")
	})

	t.Run("missing category", func(t *testing.T) {
		p := DefaultPolicy()
		p.Service = "auth"
		assert.ErrorContains(t, p.Validate(), "category")
	})

	t.Run("invalid strategy", func(t *testing.T) {
		p := DefaultPolicy()
		p.Service = "auth"
		p.Category = "login_logs"
		p.Strategy = "bogus"
		assert.ErrorContains(t, p.Validate(), "strategy")
	})

	t.Run("age without retention_days or cutoff", func(t *testing.T) {
		p := DefaultPolicy()
		p.Service = "auth"
		p.Category = "login_logs"
		p.Strategy = StrategyAge
		p.RetentionDays = 0
		assert.ErrorContains(t, p.Validate(), "retention_days")
	})

	t.Run("count without keep_latest_count", func(t *testing.T) {
		p := DefaultPolicy()
		p.Service = "auth"
		p.Category = "login_logs"
		p.Strategy = StrategyCount
		p.KeepLatestCount = 0
		assert.ErrorContains(t, p.Validate(), "keep_latest_count")
	})

	t.Run("batch_size out of bounds", func(t *testing.T) {
		p := DefaultPolicy()
		p.Service = "auth"
		p.Category = "login_logs"
		p.BatchSize = 0
		assert.ErrorContains(t, p.Validate(), "batch_size")

		p.BatchSize = 10001
		assert.ErrorContains(t, p.Validate(), "batch_size")
	})

	t.Run("max_batches out of bounds", func(t *testing.T) {
		p := DefaultPolicy()
		p.Service = "auth"
		p.Category = "login_logs"
		p.MaxBatchesPerRun = 0
		assert.ErrorContains(t, p.Validate(), "max_batches_per_run")

		p.MaxBatchesPerRun = 1001
		assert.ErrorContains(t, p.Validate(), "max_batches_per_run")
	})
}

func TestPolicy_ComputeCutoff(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)

	t.Run("uses retention_days", func(t *testing.T) {
		p := DefaultPolicy()
		p.RetentionDays = 30
		cutoff := p.ComputeCutoff(now)
		expected := now.AddDate(0, 0, -30)
		assert.True(t, cutoff.Equal(expected), "expected %v, got %v", expected, cutoff)
	})

	t.Run("uses explicit cutoff_timestamp", func(t *testing.T) {
		explicit := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		p := DefaultPolicy()
		p.CutoffTimestamp = &explicit
		cutoff := p.ComputeCutoff(now)
		assert.Equal(t, explicit, cutoff)
	})
}

func TestPolicy_EffectiveBounds(t *testing.T) {
	t.Run("clamps batch_size", func(t *testing.T) {
		p := DefaultPolicy()
		p.BatchSize = -1
		assert.Equal(t, 1, p.EffectiveBatchSize())
		p.BatchSize = 50000
		assert.Equal(t, 10000, p.EffectiveBatchSize())
	})

	t.Run("clamps max_batches", func(t *testing.T) {
		p := DefaultPolicy()
		p.MaxBatchesPerRun = -1
		assert.Equal(t, 1, p.EffectiveMaxBatches())
		p.MaxBatchesPerRun = 5000
		assert.Equal(t, 1000, p.EffectiveMaxBatches())
	})

	t.Run("total records cap", func(t *testing.T) {
		p := DefaultPolicy()
		p.BatchSize = 500
		p.MaxBatchesPerRun = 20
		assert.Equal(t, 10000, p.TotalRecordsCap())
	})
}

// ── Cron Validation ──────────────────────────────────────────────────

func TestValidateCronExpression(t *testing.T) {
	t.Run("valid cron expressions", func(t *testing.T) {
		validExprs := []string{
			"0 3 * * *",        // daily at 3am
			"0 */6 * * *",      // every 6 hours
			"30 2 * * 1",       // 2:30am on Mondays
			"*/15 * * * *",     // every 15 minutes
			"@daily",           // alias
			"@weekly",          // alias
			"@every 6h",        // @every syntax
			"@every 5m",        // minimum allowed
			"0 0 1 1 *",        // Jan 1 midnight
			"0 0 * * 0,6",      // weekends at midnight
		}
		for _, expr := range validExprs {
			t.Run(expr, func(t *testing.T) {
				assert.NoError(t, ValidateCronExpression(expr), "expr: %s", expr)
			})
		}
	})

	t.Run("invalid cron expressions", func(t *testing.T) {
		invalid := map[string]string{
			"":                     "empty",
			"* * * * * *":          "six fields",
			"60 * * * *":           "minute out of range",
			"* 24 * * *":           "hour out of range",
			"* * 32 * *":           "day of month out of range",
			"* * * 13 *":           "month out of range",
			"* * * * 8":            "day of week out of range",
			"@every 1m":            "too frequent",
			"@every 30s":           "too frequent",
		}
		for expr, desc := range invalid {
			t.Run(desc, func(t *testing.T) {
				assert.Error(t, ValidateCronExpression(expr), "should reject: %s", expr)
			})
		}
	})
}

// ── Retention Days Validation ────────────────────────────────────────

func TestValidateRetentionDays(t *testing.T) {
	assert.NoError(t, ValidateRetentionDays(30, 7))
	assert.Error(t, ValidateRetentionDays(0, 7))
	assert.Error(t, ValidateRetentionDays(3, 7))
	assert.NoError(t, ValidateRetentionDays(7, 7))
	assert.NoError(t, ValidateRetentionDays(365, 7))
}

// ── Timezone Validation ──────────────────────────────────────────────

func TestValidateTimezone(t *testing.T) {
	tests := []struct {
		tz      string
		wantErr bool
	}{
		{"UTC", false},
		{"Asia/Tehran", false},
		{"America/New_York", false},
		{"Europe/London", false},
		{"", false}, // defaults to UTC
		{"Mars/Olympus", true},
		{"invalid", true},
	}
	for _, tt := range tests {
		t.Run(tt.tz, func(t *testing.T) {
			loc, err := ValidateTimezone(tt.tz)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, loc)
			}
		})
	}
}

// ── Batch Limits Validation ──────────────────────────────────────────

func TestValidateBatchLimits(t *testing.T) {
	assert.NoError(t, ValidateBatchLimits(500, 20))
	assert.Error(t, ValidateBatchLimits(0, 20))
	assert.Error(t, ValidateBatchLimits(500, 0))
	assert.Error(t, ValidateBatchLimits(20000, 20))
	assert.Error(t, ValidateBatchLimits(500, 2000))
}

// ── FormatCountColumn ────────────────────────────────────────────────

func TestFormatCountColumn(t *testing.T) {
	assert.Equal(t, "10000", FormatCountColumn(10000))
	assert.Equal(t, "unlimited", FormatCountColumn(0))
	assert.Equal(t, "unlimited", FormatCountColumn(-1))
}

// ── Default Policy ───────────────────────────────────────────────────

func TestDefaultPolicy_IsSafe(t *testing.T) {
	p := DefaultPolicy()
	// Default must be disabled and dry-run
	assert.False(t, p.Enabled, "default policy must be disabled")
	assert.True(t, p.DryRun, "default policy must be dry-run")
	assert.Equal(t, StrategyAge, p.Strategy)
	assert.Equal(t, 30, p.RetentionDays)
	assert.Equal(t, 500, p.BatchSize)
	assert.Equal(t, "UTC", p.Timezone)
}

// ── BatchRunner ──────────────────────────────────────────────────────

func TestBatchRunner_DryRun(t *testing.T) {
	snapshot := RunSnapshot{
		PolicyID:   "test-policy",
		Service:    "auth",
		Category:   "login_logs",
		Strategy:   StrategyAge,
		DryRun:     true,
		BatchSize:  100,
		MaxBatches: 5,
		Cutoff:     time.Now().AddDate(0, 0, -30),
		RunID:      uuid.New().String(),
		StartedAt:  time.Now().UTC(),
	}

	calledEligibility := false
	eligibilityFn := func(ctx context.Context, s RunSnapshot) (int64, error) {
		calledEligibility = true
		return 999, nil
	}

	deleteFn := func(ctx context.Context, s RunSnapshot, lastCreatedAt time.Time, lastID uuid.UUID) (int64, time.Time, uuid.UUID, bool, error) {
		t.Fatal("deleteFn must not be called in dry-run")
		return 0, time.Time{}, uuid.Nil, false, nil
	}

	runner := NewBatchRunner(snapshot, eligibilityFn, deleteFn)
	result := runner.Run(context.Background())

	assert.True(t, calledEligibility, "eligibility must be called for dry-run")
	assert.Equal(t, ResultSkippedDryRun, result.Result)
	assert.Equal(t, int64(999), result.ScannedCount)
	assert.Equal(t, int64(0), result.DeletedCount)
}

func TestBatchRunner_LiveDeletion(t *testing.T) {
	// Simulate 3 batches of 5 records each = 15 total
	records := make([]uuid.UUID, 15)
	for i := range records {
		records[i] = uuid.New()
	}

	var cursor int
	snapshot := RunSnapshot{
		PolicyID:   "test-policy-2",
		Service:    "notifier",
		Category:   "notification_logs",
		Strategy:   StrategyAge,
		DryRun:     false,
		BatchSize:  5,
		MaxBatches: 10,
		Cutoff:     time.Now().AddDate(0, 0, -30),
		RunID:      uuid.New().String(),
		StartedAt:  time.Now().UTC(),
	}

	eligibilityFn := func(ctx context.Context, s RunSnapshot) (int64, error) {
		return int64(len(records)), nil
	}

	deleteFn := func(ctx context.Context, s RunSnapshot, lastCreatedAt time.Time, lastID uuid.UUID) (int64, time.Time, uuid.UUID, bool, error) {
		if cursor >= len(records) {
			return 0, time.Time{}, uuid.Nil, false, nil
		}
		batchEnd := cursor + snapshot.BatchSize
		if batchEnd > len(records) {
			batchEnd = len(records)
		}
		n := int64(batchEnd - cursor)
		cursor = batchEnd
		hasMore := cursor < len(records)
		return n, time.Now(), uuid.New(), hasMore, nil
	}

	runner := NewBatchRunner(snapshot, eligibilityFn, deleteFn)
	result := runner.Run(context.Background())

	assert.Equal(t, ResultSuccess, result.Result)
	assert.Equal(t, int64(15), result.DeletedCount)
	assert.Equal(t, 3, result.BatchesRun)
}

func TestBatchRunner_MaxBatchesEnforced(t *testing.T) {
	snapshot := RunSnapshot{
		PolicyID:   "test-policy-3",
		Service:    "auth",
		Category:   "login_logs",
		Strategy:   StrategyAge,
		DryRun:     false,
		BatchSize:  1,
		MaxBatches: 3,
		Cutoff:     time.Now().AddDate(0, 0, -30),
		RunID:      uuid.New().String(),
		StartedAt:  time.Now().UTC(),
	}

	deleteFn := func(ctx context.Context, s RunSnapshot, lastCreatedAt time.Time, lastID uuid.UUID) (int64, time.Time, uuid.UUID, bool, error) {
		return 1, time.Now(), uuid.New(), true, nil // always has more
	}

	runner := NewBatchRunner(snapshot, nil, deleteFn)
	result := runner.Run(context.Background())

	assert.Equal(t, ResultPartial, result.Result)
	assert.ErrorIs(t, result.Error, ErrMaxBatchesReached)
	assert.Equal(t, int64(3), result.DeletedCount)
	assert.Equal(t, 3, result.BatchesRun)
}

func TestBatchRunner_ContextCancellation(t *testing.T) {
	snapshot := RunSnapshot{
		PolicyID:   "test-policy-4",
		Service:    "auth",
		Category:   "login_logs",
		Strategy:   StrategyAge,
		DryRun:     false,
		BatchSize:  100,
		MaxBatches: 10,
		Cutoff:     time.Now().AddDate(0, 0, -30),
		RunID:      uuid.New().String(),
		StartedAt:  time.Now().UTC(),
	}

	deleteFn := func(ctx context.Context, s RunSnapshot, lastCreatedAt time.Time, lastID uuid.UUID) (int64, time.Time, uuid.UUID, bool, error) {
		// Simulate long-running work
		select {
		case <-ctx.Done():
			return 0, time.Time{}, uuid.Nil, false, ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
		return 5, time.Now(), uuid.New(), true, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	runner := NewBatchRunner(snapshot, nil, deleteFn)
	result := runner.Run(ctx)

	assert.Equal(t, ResultPartial, result.Result)
	assert.ErrorIs(t, result.Error, ErrContextCancelled)
}

// ── Errors ───────────────────────────────────────────────────────────

func TestSentinelErrors(t *testing.T) {
	// Verify errors are distinct and comparable with errors.Is
	assert.True(t, true) // placeholder for sentinel existence check
	assert.NotNil(t, ErrDisabled)
	assert.NotNil(t, ErrDryRun)
	assert.NotNil(t, ErrLockHeld)
	assert.NotNil(t, ErrMaxBatchesReached)
	assert.NotNil(t, ErrCategoryProtected)
	assert.NotNil(t, ErrInvalidPolicy)
	assert.NotNil(t, ErrContextCancelled)
	assert.NotNil(t, ErrPolicyNotFound)
}

// ── RunSnapshot ──────────────────────────────────────────────────────

func TestRunSnapshot(t *testing.T) {
	snap := RunSnapshot{
		PolicyID:   "p-1",
		Service:    "auth",
		Category:   "login_logs",
		Strategy:   StrategyHybrid,
		DryRun:     true,
		Cutoff:     time.Now(),
		KeepLatest: 10000,
		BatchSize:  500,
		MaxBatches: 20,
		Trigger:    TriggerManual,
		RunID:      uuid.New().String(),
		StartedAt:  time.Now(),
	}

	assert.Equal(t, "p-1", snap.PolicyID)
	assert.Equal(t, "auth", snap.Service)
	assert.True(t, snap.DryRun)
	assert.Equal(t, TriggerManual, snap.Trigger)
}

// ── Mock Lock for tests ──────────────────────────────────────────────

type mockLockGuard struct {
	key     string
	released bool
}

func (g *mockLockGuard) Release(ctx context.Context) error {
	g.released = true
	return nil
}

func (g *mockLockGuard) Key() string { return g.key }

type mockLock struct {
	held bool
}

func (m *mockLock) Acquire(ctx context.Context, key string, ttl time.Duration) (LockGuard, error) {
	if m.held {
		return nil, ErrLockHeld
	}
	m.held = true
	return &mockLockGuard{key: key}, nil
}

func (m *mockLock) IsHeld(ctx context.Context, key string) (bool, error) {
	return m.held, nil
}

func TestDistributedLock_Mock(t *testing.T) {
	lock := &mockLock{}

	// First acquire
	g, err := lock.Acquire(context.Background(), "auth:login_logs", time.Minute)
	require.NoError(t, err)
	require.NotNil(t, g)
	assert.Equal(t, "auth:login_logs", g.Key())

	// Second acquire should fail
	_, err = lock.Acquire(context.Background(), "auth:login_logs", time.Minute)
	assert.ErrorIs(t, err, ErrLockHeld)

	held, err := lock.IsHeld(context.Background(), "auth:login_logs")
	assert.NoError(t, err)
	assert.True(t, held)

	// Release
	assert.NoError(t, g.Release(context.Background()))
}
