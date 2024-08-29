package service

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/discard"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	// MetricsSubsystem is a subsystem shared by all metrics exposed by this package.
	MetricsSubsystem = "service"
)

// Metrics contains metrics exposed by this package.
type Metrics struct {
	// How many requests for block events the service received, with success status.
	BlockEventsRequestsTotal metrics.Counter
}

func newMetrics(
	blockEventsRequestsTotal metrics.Counter,
) *Metrics {
	return &Metrics{
		BlockEventsRequestsTotal: blockEventsRequestsTotal,
	}
}

func (m *Metrics) setStartingValues() {}

func PrometheusMetrics(namespace string, labelsAndValues ...string) *Metrics {
	labels := []string{}
	for i := 0; i < len(labelsAndValues); i += 2 {
		labels = append(labels, labelsAndValues[i])
	}
	m := newMetrics(
		prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "block_events_requests_total",
			Help:      "How many requests for block events the service received, with success status.",
		}, append(labels, "success")).With(labelsAndValues...),
	)

	m.setStartingValues()
	return m
}

func NoopMetrics() *Metrics {
	return newMetrics(
		discard.NewCounter(),
	)
}
