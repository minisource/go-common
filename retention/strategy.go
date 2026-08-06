package retention

// Strategy defines how log records are selected for deletion.
type Strategy string

const (
	// StrategyAge deletes records older than a configured retention period.
	StrategyAge Strategy = "age"

	// StrategyCount keeps only the newest N records and deletes the rest.
	StrategyCount Strategy = "count"

	// StrategyHybrid applies both age and count limits.
	// Records are eligible only when they violate BOTH limits (AND semantics).
	// This is the conservative default: a record must be old enough AND beyond
	// the count threshold before it is deleted. Never use OR semantics here
	// without an explicit override flag.
	StrategyHybrid Strategy = "hybrid"
)

// ValidStrategies returns all valid strategy values.
func ValidStrategies() []Strategy {
	return []Strategy{StrategyAge, StrategyCount, StrategyHybrid}
}

// IsValid reports whether s is a recognised strategy.
func (s Strategy) IsValid() bool {
	switch s {
	case StrategyAge, StrategyCount, StrategyHybrid:
		return true
	default:
		return false
	}
}

// Trigger describes what initiated a cleanup run.
type Trigger string

const (
	TriggerScheduled Trigger = "scheduled"
	TriggerManual    Trigger = "manual"
)

// Result summarises the outcome of a cleanup run.
type Result string

const (
	ResultSuccess          Result = "success"
	ResultPartial          Result = "partial"
	ResultFailed           Result = "failed"
	ResultSkippedLockHeld  Result = "skipped_lock_held"
	ResultSkippedDisabled  Result = "skipped_disabled"
	ResultSkippedDryRun    Result = "skipped_dry_run"
)
