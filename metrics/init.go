package metrics

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// registerOrReuse registers c with reg and returns the collector that
// instrumentation must observe. If a collector with the same descriptor is
// already registered (repeated initialization, hot-reload, or a shared
// registry), the already-registered collector instance is returned so that
// observations keep updating the exported series instead of an unregistered
// duplicate.
func registerOrReuse(reg prometheus.Registerer, c prometheus.Collector) prometheus.Collector {
	if err := reg.Register(c); err != nil {
		var are prometheus.AlreadyRegisteredError
		if errors.As(err, &are) {
			return are.ExistingCollector
		}
		panic(err)
	}
	return c
}

// RegisterOrReuse registers c with the default registry and returns the
// collector to use: the newly registered collector, or an equivalent collector
// that was already registered.
func RegisterOrReuse(c prometheus.Collector) prometheus.Collector {
	return registerOrReuse(prometheus.DefaultRegisterer, c)
}

func mustCounterVec(c prometheus.Collector) *prometheus.CounterVec {
	v, ok := c.(*prometheus.CounterVec)
	if !ok {
		panic(fmt.Sprintf("metrics: expected *prometheus.CounterVec, got %T", c))
	}
	return v
}

func mustHistogramVec(c prometheus.Collector) *prometheus.HistogramVec {
	v, ok := c.(*prometheus.HistogramVec)
	if !ok {
		panic(fmt.Sprintf("metrics: expected *prometheus.HistogramVec, got %T", c))
	}
	return v
}

// InitMetrics registers all metrics with Prometheus. It is safe to call
// multiple times: repeat registrations reuse the already-registered collector
// instance, and the package-level collectors are rebound to that instance so
// all observations land on the exported series.
func InitMetrics() {
	HttpDuration = mustHistogramVec(RegisterOrReuse(HttpDuration))
	HttpRequestsTotal = mustCounterVec(RegisterOrReuse(HttpRequestsTotal))
	DbCall = mustCounterVec(RegisterOrReuse(DbCall))
	DbQueryDuration = mustHistogramVec(RegisterOrReuse(DbQueryDuration))
	CacheHitsTotal = mustCounterVec(RegisterOrReuse(CacheHitsTotal))
	CacheMissesTotal = mustCounterVec(RegisterOrReuse(CacheMissesTotal))
}