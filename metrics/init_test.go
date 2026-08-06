package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRegisterOrReuse_FirstCallRegistersGivenCollector(t *testing.T) {
	reg := prometheus.NewRegistry()
	c := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_first_counter", Help: "h"}, []string{"code"})

	got := registerOrReuse(reg, c)
	if got != c {
		t.Fatalf("first registration must return the given collector, got %T", got)
	}
}

func TestRegisterOrReuse_RepeatedRegistrationReturnsSameCollectorAndGathers(t *testing.T) {
	reg := prometheus.NewRegistry()
	c := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_repeat_counter", Help: "h"}, []string{"code"})

	first := registerOrReuse(reg, c)
	second := registerOrReuse(reg, c)
	if first != c || second != c {
		t.Fatalf("repeated registration must keep returning the registered collector")
	}

	// Observations through the returned collector must be visible to Gather.
	second.(*prometheus.CounterVec).WithLabelValues("200").Inc()
	second.(*prometheus.CounterVec).WithLabelValues("200").Inc()

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	var total float64
	for _, mf := range mfs {
		for _, m := range mf.GetMetric() {
			if m.GetCounter() != nil {
				total += m.GetCounter().GetValue()
			}
		}
	}
	if total != 2 {
		t.Fatalf("expected gathered counter value 2, got %v", total)
	}
}

func TestRegisterOrReuse_ReturnsExistingRegisteredCollector(t *testing.T) {
	reg := prometheus.NewRegistry()
	existing := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_existing_counter", Help: "h"}, []string{"l"})
	if err := reg.Register(existing); err != nil {
		t.Fatalf("seed register: %v", err)
	}

	// A different instance with an identical descriptor must not be registered;
	// the already-registered collector is returned instead.
	replacement := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_existing_counter", Help: "h"}, []string{"l"})
	got := registerOrReuse(reg, replacement)
	if got != existing {
		t.Fatalf("expected the already-registered collector, got a different instance")
	}

	// Observing through the returned existing collector must update the
	// exported series.
	got.(*prometheus.CounterVec).WithLabelValues("x").Inc()
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	val := mfs[0].GetMetric()[0].GetCounter().GetValue()
	if val != 1 {
		t.Fatalf("expected gathered value 1, got %v", val)
	}
}

func TestRegisterOrReuse_IndependentRegistriesAreIsolated(t *testing.T) {
	regA := prometheus.NewRegistry()
	regB := prometheus.NewRegistry()
	newCounter := func() *prometheus.CounterVec {
		return prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_iso_counter", Help: "h"}, nil)
	}

	a := registerOrReuse(regA, newCounter()).(*prometheus.CounterVec)
	b := registerOrReuse(regB, newCounter()).(*prometheus.CounterVec)

	a.WithLabelValues().Inc()
	b.WithLabelValues().Inc()
	b.WithLabelValues().Inc()

	mfsA, _ := regA.Gather()
	mfsB, _ := regB.Gather()
	valA := mfsA[0].GetMetric()[0].GetCounter().GetValue()
	valB := mfsB[0].GetMetric()[0].GetCounter().GetValue()
	if valA != 1 || valB != 2 {
		t.Fatalf("registries not isolated: A=%v B=%v", valA, valB)
	}
}

func TestInitMetrics_RepeatedCallsAreIdempotent(t *testing.T) {
	// Repeated initialization on the default registry must not panic, and the
	// package-level collectors must keep observing the exported series.
	InitMetrics()
	InitMetrics()
	InitMetrics()

	HttpRequestsTotal.WithLabelValues("/probe", "GET", "200").Inc()
	if got := testutil.ToFloat64(HttpRequestsTotal.WithLabelValues("/probe", "GET", "200")); got != 1 {
		t.Fatalf("expected counter value 1 after repeated InitMetrics, got %v", got)
	}

	HttpDuration.WithLabelValues("/probe", "GET", "200").Observe(3)
	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	found := false
	for _, mf := range mfs {
		if mf.GetName() != "http_response_time" {
			continue
		}
		for _, m := range mf.GetMetric() {
			if m.GetHistogram() == nil {
				continue
			}
			labels := map[string]string{}
			for _, l := range m.GetLabel() {
				labels[l.GetName()] = l.GetValue()
			}
			if labels["path"] == "/probe" && labels["method"] == "GET" {
				found = true
				if got := m.GetHistogram().GetSampleSum(); got != 3 {
					t.Fatalf("expected histogram sum 3 after repeated InitMetrics, got %v", got)
				}
			}
		}
	}
	if !found {
		t.Fatalf("probe series not found in default gatherer after repeated InitMetrics")
	}
}
