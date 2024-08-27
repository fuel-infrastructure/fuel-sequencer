package sidecar

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/discard"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	// MetricsSubsystem is a subsystem shared by all metrics exposed by this package.
	MetricsSubsystem = "sidecar"
)

// Metrics contains metrics exposed by this package.
type Metrics struct {
	// Whether the sidecar is catching up to Ethereum. 1 if yes, 0 if no.
	CatchingUp metrics.Gauge
	// The last Ethereum block that the Sidecar can sync at the moment.
	MaxSyncableBlock metrics.Gauge
}

func (m *Metrics) setStartingValues() {
	m.CatchingUp.Set(0) // not catching up by default
}

func PrometheusMetrics(namespace string, labelsAndValues ...string) *Metrics {
	labels := []string{}
	for i := 0; i < len(labelsAndValues); i += 2 {
		labels = append(labels, labelsAndValues[i])
	}
	m := &Metrics{
		CatchingUp: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "catching_up",
			Help:      "Whether the sidecar is catching up to ethereum. 1 if yes, 0 if no.",
		}, labels).With(labelsAndValues...),
		MaxSyncableBlock: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "max_syncable_block",
			Help:      "The last Ethereum block that the Sidecar can sync at the moment.",
		}, labels).With(labelsAndValues...),
	}

	m.setStartingValues()
	return m
}

func NopMetrics() *Metrics {
	m := &Metrics{
		CatchingUp:       discard.NewGauge(),
		MaxSyncableBlock: discard.NewGauge(),
	}

	m.setStartingValues()
	return m
}
