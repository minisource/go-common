package retention

import (
	"github.com/minisource/go-common/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// Retention-metrics package-level collectors. These are registered once via
// InitMetrics and then safely reused across hot-reloads.
var (
	CleanupRunsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "log_cleanup_runs_total",
		Help: "Total number of log cleanup runs.",
	}, []string{"service", "category", "result", "trigger"})

	CleanupDeletedRecordsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "log_cleanup_deleted_records_total",
		Help: "Total number of records deleted by log cleanup.",
	}, []string{"service", "category"})

	CleanupScannedRecordsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "log_cleanup_scanned_records_total",
		Help: "Total number of records scanned (preview count) by log cleanup.",
	}, []string{"service", "category"})

	CleanupDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "log_cleanup_duration_seconds",
		Help:    "Duration of log cleanup runs.",
		Buckets: []float64{0.1, 0.5, 1, 5, 10, 30, 60, 300, 600, 1800, 3600},
	}, []string{"service", "category", "result"})

	CleanupActive = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "log_cleanup_active",
		Help: "Whether a cleanup run is currently active (1) or not (0).",
	}, []string{"service", "category"})

	CleanupLastSuccessTimestamp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "log_cleanup_last_success_timestamp_seconds",
		Help: "Unix timestamp of the last successful cleanup run.",
	}, []string{"service", "category"})

	CleanupFailuresTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "log_cleanup_failures_total",
		Help: "Total number of failed cleanup runs.",
	}, []string{"service", "category", "reason"})

	CleanupLockContentionTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "log_cleanup_lock_contention_total",
		Help: "Total number of times a cleanup run was skipped because the lock was held.",
	}, []string{"service", "category"})
)

// init registers retention metrics with the default Prometheus registry using
// safe reuse semantics (compatible with hot-reload / air).
func init() {
	// Overwrite package-level collectors with the actual registered instances
	// so that observations always land on the exported series.
	CleanupRunsTotal = mustCounterVec(metrics.RegisterOrReuse(CleanupRunsTotal))
	CleanupDeletedRecordsTotal = mustCounterVec(metrics.RegisterOrReuse(CleanupDeletedRecordsTotal))
	CleanupScannedRecordsTotal = mustCounterVec(metrics.RegisterOrReuse(CleanupScannedRecordsTotal))
	CleanupDurationSeconds = mustHistogramVec(metrics.RegisterOrReuse(CleanupDurationSeconds))
	CleanupActive = mustGaugeVec(metrics.RegisterOrReuse(CleanupActive))
	CleanupLastSuccessTimestamp = mustGaugeVec(metrics.RegisterOrReuse(CleanupLastSuccessTimestamp))
	CleanupFailuresTotal = mustCounterVec(metrics.RegisterOrReuse(CleanupFailuresTotal))
	CleanupLockContentionTotal = mustCounterVec(metrics.RegisterOrReuse(CleanupLockContentionTotal))
}

func mustCounterVec(c prometheus.Collector) *prometheus.CounterVec {
	v, ok := c.(*prometheus.CounterVec)
	if !ok {
		panic("metrics: retention counter is not *prometheus.CounterVec")
	}
	return v
}

func mustHistogramVec(c prometheus.Collector) *prometheus.HistogramVec {
	v, ok := c.(*prometheus.HistogramVec)
	if !ok {
		panic("metrics: retention histogram is not *prometheus.HistogramVec")
	}
	return v
}

func mustGaugeVec(c prometheus.Collector) *prometheus.GaugeVec {
	v, ok := c.(*prometheus.GaugeVec)
	if !ok {
		panic("metrics: retention gauge is not *prometheus.GaugeVec")
	}
	return v
}
