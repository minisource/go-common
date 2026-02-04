package metrics

import "github.com/prometheus/client_golang/prometheus"

var DbCall = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "db_calls_total",
		Help: "Number of database calls",
	}, []string{"type_name", "operation_name", "status"},
)

var HttpRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"path", "method", "status_code"},
)

var CacheHitsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "Total number of cache hits",
	}, []string{"cache_type"},
)

var CacheMissesTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "Total number of cache misses",
	}, []string{"cache_type"},
)
