package store

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/discard"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	// MetricsSubsystem is a subsystem shared by all metrics exposed by this package.
	MetricsSubsystem = "store"
)

// Metrics contains metrics exposed by this package.
type Metrics struct {
	// The last Ethereum block synced by the Sidecar.
	LastSyncedBlock metrics.Gauge
	// The oldest Ethereum block that the Sidecar has in state, if any
	StartQueryBlock metrics.Gauge
	// The number of events processed by the Sidecar.
	EventsProcessed metrics.Counter
	// The number of old Ethereum blocks pruned by the Sidecar.
	BlocksPruned metrics.Counter
}

func (m *Metrics) setStartingValues() {}

func PrometheusMetrics(namespace string, labelsAndValues ...string) *Metrics {
	labels := []string{}
	for i := 0; i < len(labelsAndValues); i += 2 {
		labels = append(labels, labelsAndValues[i])
	}
	m := &Metrics{
		LastSyncedBlock: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "last_synced_block",
			Help:      "The last Ethereum block synced by the Sidecar.",
		}, labels).With(labelsAndValues...),
		StartQueryBlock: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "start_query_block",
			Help:      "The oldest Ethereum block that the Sidecar has in state, if any.",
		}, labels).With(labelsAndValues...),
		EventsProcessed: prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "events_processed",
			Help:      "The number of events processed by the Sidecar.",
		}, labels).With(labelsAndValues...),
		BlocksPruned: prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "blocks_pruned",
			Help:      "The number of old Ethereum blocks pruned by the Sidecar.",
		}, labels).With(labelsAndValues...),
	}

	m.setStartingValues()
	return m
}

func NopMetrics() *Metrics {
	m := &Metrics{
		LastSyncedBlock: discard.NewGauge(),
		StartQueryBlock: discard.NewGauge(),
		EventsProcessed: discard.NewCounter(),
		BlocksPruned:    discard.NewCounter(),
	}

	m.setStartingValues()
	return m
}
