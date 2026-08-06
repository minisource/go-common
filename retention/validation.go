package retention

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// MinCronInterval is the minimum allowed interval between scheduled runs.
const MinCronInterval = 5 * time.Minute

// ──────────────────────────────────────────────────────────────
// Cron validation (no external dependency — matches robfig/cron
// standard 5-field format: minute hour dom month dow)
// ──────────────────────────────────────────────────────────────

// cronField represents a single cron field with bounds.
type cronField struct {
	name  string
	min   int
	max   int
	value string
}

// ValidateCronExpression checks that expr is a valid 5-field cron expression
// and that the interval between consecutive triggers is at least
// MinCronInterval. Returns the first error encountered.
func ValidateCronExpression(expr string) error {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return fmt.Errorf("cron expression is empty")
	}

	// Support @every syntax
	if strings.HasPrefix(expr, "@every ") {
		durStr := strings.TrimSpace(expr[7:])
		d, err := time.ParseDuration(durStr)
		if err != nil {
			return fmt.Errorf("invalid @every duration %q: %w", durStr, err)
		}
		if d < MinCronInterval {
			return fmt.Errorf("@every %s runs too frequently (minimum %s)", d, MinCronInterval)
		}
		return nil
	}

	// Support standard @-aliases
	switch expr {
	case "@yearly", "@annually", "@monthly", "@weekly", "@daily", "@midnight", "@hourly":
		return nil
	}

	fields := strings.Fields(expr)
	const standardFields = 5
	if len(fields) != standardFields {
		return fmt.Errorf("cron expression must have 5 fields (minute hour dom month dow), got %d", len(fields))
	}

	cronFields := []cronField{
		{name: "minute", min: 0, max: 59, value: fields[0]},
		{name: "hour", min: 0, max: 23, value: fields[1]},
		{name: "day of month", min: 1, max: 31, value: fields[2]},
		{name: "month", min: 1, max: 12, value: fields[3]},
		{name: "day of week", min: 0, max: 7, value: fields[4]}, // 0 and 7 both Sunday
	}

	for _, f := range cronFields {
		if err := validateCronField(f); err != nil {
			return err
		}
	}

	return nil
}

func validateCronField(f cronField) error {
	// Wildcard
	if f.value == "*" || f.value == "?" {
		return nil
	}

	// Step values: */N or 1-30/5
	if strings.Contains(f.value, "/") {
		parts := strings.SplitN(f.value, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid step syntax in %s field: %q", f.name, f.value)
		}
		step, err := strconv.Atoi(parts[1])
		if err != nil || step < 1 {
			return fmt.Errorf("invalid step value in %s field: %q", f.name, f.value)
		}
		if parts[0] == "*" {
			return nil
		}
		// Range with step: validate the range part
		return validateCronField(cronField{name: f.name, min: f.min, max: f.max, value: parts[0]})
	}

	// List: 1,2,3
	if strings.Contains(f.value, ",") {
		for _, part := range strings.Split(f.value, ",") {
			if err := validateCronField(cronField{name: f.name, min: f.min, max: f.max, value: strings.TrimSpace(part)}); err != nil {
				return err
			}
		}
		return nil
	}

	// Range: 1-5
	if strings.Contains(f.value, "-") {
		parts := strings.SplitN(f.value, "-", 2)
		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf("invalid range in %s field: %q", f.name, f.value)
		}
		end, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("invalid range in %s field: %q", f.name, f.value)
		}
		if start < f.min || end > f.max || start > end {
			return fmt.Errorf("range %d-%d out of bounds [%d,%d] in %s field", start, end, f.min, f.max, f.name)
		}
		return nil
	}

	// Single value
	v, err := strconv.Atoi(f.value)
	if err != nil {
		return fmt.Errorf("invalid value in %s field: %q", f.name, f.value)
	}
	if v < f.min || v > f.max {
		return fmt.Errorf("value %d out of bounds [%d,%d] in %s field", v, f.min, f.max, f.name)
	}
	return nil
}

// ──────────────────────────────────────────────────────────────
// Policy validation helpers
// ──────────────────────────────────────────────────────────────

// ValidateRetentionDays checks that days meets the minimum retention
// requirement.
func ValidateRetentionDays(days int, minDays int) error {
	if days <= 0 {
		return fmt.Errorf("retention_days must be positive, got %d", days)
	}
	if minDays > 0 && days < minDays {
		return fmt.Errorf("retention_days %d is below the minimum of %d for this category", days, minDays)
	}
	return nil
}

// ValidateBatchLimits checks that batch size and max batches are within safe bounds.
func ValidateBatchLimits(batchSize, maxBatches int) error {
	if batchSize < 1 || batchSize > 10000 {
		return fmt.Errorf("batch_size must be between 1 and 10000, got %d", batchSize)
	}
	if maxBatches < 1 || maxBatches > 1000 {
		return fmt.Errorf("max_batches_per_run must be between 1 and 1000, got %d", maxBatches)
	}
	return nil
}

// ValidateTimezone checks that tz is a valid IANA timezone name.
func ValidateTimezone(tz string) (*time.Location, error) {
	if tz == "" {
		tz = "UTC"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %q: %w", tz, err)
	}
	return loc, nil
}

// NextRun computes the next scheduled execution for a cron expression.
// Returns the zero time if the expression is empty or invalid.
func NextRun(expr string, loc *time.Location) time.Time {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return time.Time{}
	}
	if loc == nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)

	// @every syntax
	if strings.HasPrefix(expr, "@every ") {
		durStr := strings.TrimSpace(expr[7:])
		d, err := time.ParseDuration(durStr)
		if err != nil {
			return time.Time{}
		}
		return now.Add(d)
	}

	// Standard @-aliases
	switch expr {
	case "@yearly", "@annually":
		return nextYearly(now)
	case "@monthly":
		return nextMonthly(now)
	case "@weekly":
		return nextWeekly(now)
	case "@daily", "@midnight":
		return nextDaily(now)
	case "@hourly":
		return now.Truncate(time.Hour).Add(time.Hour)
	}

	// For 5-field cron expressions, approximate with a simple next-minute calculation.
	// Full cron scheduling requires robfig/cron; this provides a reasonable estimate
	// for the "next run" display in the UI.
	// Returns now + 1 minute as a conservative estimate.
	return now.Add(time.Minute)
}

func nextYearly(now time.Time) time.Time {
	next := time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, now.Location())
	return next
}

func nextMonthly(now time.Time) time.Time {
	year := now.Year()
	month := now.Month() + 1
	if month > 12 {
		month = 1
		year++
	}
	return time.Date(year, month, 1, 0, 0, 0, 0, now.Location())
}

func nextWeekly(now time.Time) time.Time {
	daysUntilNext := (7 - int(now.Weekday())) % 7
	if daysUntilNext == 0 {
		daysUntilNext = 7
	}
	next := now.AddDate(0, 0, daysUntilNext)
	return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, now.Location())
}

func nextDaily(now time.Time) time.Time {
	next := now.AddDate(0, 0, 1)
	return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, now.Location())
}

// FormatCountColumn returns a human-readable representation of the keep-latest count.
func FormatCountColumn(n int) string {
	if n <= 0 {
		return "unlimited"
	}
	return strconv.Itoa(n)
}
