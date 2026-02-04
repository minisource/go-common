package metrics

import "github.com/prometheus/client_golang/prometheus"

var HttpDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_response_time",
		Help:    "Duration of HTTP requests",
		Buckets: []float64{1, 2, 5, 10, 50, 100, 200, 500, 1000, 2000, 5000, 10000},
	}, []string{"path", "method", "status_code"})

var DbQueryDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "db_query_duration_milliseconds",
		Help:    "Duration of database queries in milliseconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation", "table"})
