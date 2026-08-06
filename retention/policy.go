package retention

import (
	"fmt"
	"time"
)

// Policy defines a log retention policy for a specific service and category.
// It is a domain model — not a database entity. Persistence is handled by
// each consuming service.
type Policy struct {
	ID          string   // Unique policy identifier (UUID string)
	Service     string   // "auth" or "notifier"
	Category    string   // Log category (e.g. "login_logs", "notification_logs")
	Enabled     bool     // Master switch — must be true for any deletion
	Strategy    Strategy // age | count | hybrid
	Description string   // Human-readable description

	// Age strategy parameters
	RetentionDays    int       // Delete records older than this many days
	CutoffTimestamp  *time.Time // Explicit cutoff (overrides RetentionDays when set)

	// Count strategy parameters
	KeepLatestCount int // Keep only the newest N records

	// Execution parameters
	CronExpression    string // Standard cron expression (5-field)
	Timezone          string // IANA timezone, default "UTC"
	BatchSize         int    // Records per batch
	MaxBatchesPerRun  int    // Safety cap on batches per execution
	DryRun            bool   // When true, preview only — never delete
	MinRetentionDays  int    // Safety floor (hardcoded per category, not user-configurable)

	// Audit
	CreatedAt time.Time
	UpdatedAt time.Time
	UpdatedBy string
	LastRunAt *time.Time
	NextRunAt *time.Time
}

// DefaultPolicy returns a safe, disabled policy with conservative defaults.
func DefaultPolicy() Policy {
	return Policy{
		Enabled:          false,
		Strategy:         StrategyAge,
		RetentionDays:    30,
		KeepLatestCount:  100000,
		BatchSize:        500,
		MaxBatchesPerRun: 20,
		DryRun:           true,
		Timezone:         "UTC",
		MinRetentionDays: 7,
	}
}

// Validate performs domain-level validation of the policy.
// It returns nil when the policy is valid for execution.
func (p *Policy) Validate() error {
	if p.Service == "" {
		return fmt.Errorf("service is required")
	}
	if p.Category == "" {
		return fmt.Errorf("category is required")
	}
	if !p.Strategy.IsValid() {
		return fmt.Errorf("invalid strategy: %s", p.Strategy)
	}

	switch p.Strategy {
	case StrategyAge:
		if p.RetentionDays <= 0 && p.CutoffTimestamp == nil {
			return fmt.Errorf("age strategy requires retention_days > 0 or cutoff_timestamp")
		}
	case StrategyCount:
		if p.KeepLatestCount <= 0 {
			return fmt.Errorf("count strategy requires keep_latest_count > 0")
		}
	case StrategyHybrid:
		if p.RetentionDays <= 0 && p.CutoffTimestamp == nil {
			return fmt.Errorf("hybrid strategy requires age limit (retention_days or cutoff_timestamp)")
		}
		if p.KeepLatestCount <= 0 {
			return fmt.Errorf("hybrid strategy requires keep_latest_count > 0")
		}
	}

	if p.BatchSize <= 0 || p.BatchSize > 10000 {
		return fmt.Errorf("batch_size must be between 1 and 10000, got %d", p.BatchSize)
	}
	if p.MaxBatchesPerRun <= 0 || p.MaxBatchesPerRun > 1000 {
		return fmt.Errorf("max_batches_per_run must be between 1 and 1000, got %d", p.MaxBatchesPerRun)
	}

	return nil
}

// ComputeCutoff returns the age cutoff time. If CutoffTimestamp is set it is
// returned; otherwise RetentionDays is subtracted from now.
func (p *Policy) ComputeCutoff(now time.Time) time.Time {
	if p.CutoffTimestamp != nil {
		return *p.CutoffTimestamp
	}
	return now.AddDate(0, 0, -p.RetentionDays)
}

// EffectiveBatchSize returns BatchSize clamped to [1, 10000].
func (p *Policy) EffectiveBatchSize() int {
	if p.BatchSize < 1 {
		return 1
	}
	if p.BatchSize > 10000 {
		return 10000
	}
	return p.BatchSize
}

// EffectiveMaxBatches returns MaxBatchesPerRun clamped to [1, 1000].
func (p *Policy) EffectiveMaxBatches() int {
	if p.MaxBatchesPerRun < 1 {
		return 1
	}
	if p.MaxBatchesPerRun > 1000 {
		return 1000
	}
	return p.MaxBatchesPerRun
}

// TotalRecordsCap returns the maximum records that can be deleted in one run.
func (p *Policy) TotalRecordsCap() int {
	return p.EffectiveBatchSize() * p.EffectiveMaxBatches()
}

// RunSnapshot captures the policy state at the start of a run so that
// concurrent policy updates do not affect an in-flight cleanup.
type RunSnapshot struct {
	PolicyID     string
	Service      string
	Category     string
	Strategy     Strategy
	DryRun       bool
	Cutoff       time.Time // Frozen at run start (for age/hybrid)
	KeepLatest   int       // Frozen at run start (for count/hybrid)
	BatchSize    int
	MaxBatches   int
	Trigger      Trigger
	RunID        string
	StartedAt    time.Time
}
