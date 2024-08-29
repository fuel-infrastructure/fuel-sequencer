package ethwrappedclient

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/discard"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	// MetricsSubsystem is a subsystem shared by all metrics exposed by this package.
	MetricsSubsystem = "eth_client"
)

// Metrics contains metrics exposed by this package.
type Metrics struct {
	// Histogram of how long it takes to receive queried Ethereum logs.
	LogsQueryResponseTimeSeconds metrics.Histogram
}

func (m *Metrics) setStartingValues() {}

func PrometheusMetrics(namespace string, labelsAndValues ...string) *Metrics {
	labels := []string{}
	for i := 0; i < len(labelsAndValues); i += 2 {
		labels = append(labels, labelsAndValues[i])
	}
	m := &Metrics{
		LogsQueryResponseTimeSeconds: prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "logs_query_response_time_seconds",
			Help:      "Histogram of how long it takes to receive queried Ethereum logs.",

			Buckets: stdprometheus.ExponentialBucketsRange(0.5, 30, 8),
		}, labels).With(labelsAndValues...),
	}

	m.setStartingValues()
	return m
}

func NopMetrics() *Metrics {
	m := &Metrics{
		LogsQueryResponseTimeSeconds: discard.NewHistogram(),
	}

	m.setStartingValues()
	return m
}
