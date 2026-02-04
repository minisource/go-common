package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// InitMetrics registers all metrics with Prometheus
func InitMetrics() {
	// Register HTTP metrics
	prometheus.MustRegister(HttpDuration)
	prometheus.MustRegister(HttpRequestsTotal)

	// Register DB metrics
	prometheus.MustRegister(DbCall)
	prometheus.MustRegister(DbQueryDuration)

	// Register cache metrics
	prometheus.MustRegister(CacheHitsTotal)
	prometheus.MustRegister(CacheMissesTotal)
}
