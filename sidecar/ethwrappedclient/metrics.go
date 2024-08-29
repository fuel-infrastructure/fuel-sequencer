package ethwrappedclient

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/discard"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	// MetricsSubsystem is a subsystem shared by all metrics exposed by this package.
	MetricsSubsystem = "eth"
)

// Metrics contains metrics exposed by this package.
type Metrics struct {
	// Histogram of how long it takes to receive queried Ethereum logs.
	LogsQueryDelaySeconds metrics.Histogram
	// Counter of how many errors were observed when querying Ethereum logs.
	LogsQueryErrorCount metrics.Counter
}

func newMetrics(
	logsQueryDelaySeconds metrics.Histogram,
	logsQueryErrorCount metrics.Counter,
) *Metrics {
	return &Metrics{
		LogsQueryDelaySeconds: logsQueryDelaySeconds,
		LogsQueryErrorCount:   logsQueryErrorCount,
	}
}

func (m *Metrics) setStartingValues() {}

func (m *Metrics) ObserveLogsQueryDelay(from, to time.Time) {
	m.LogsQueryDelaySeconds.Observe(to.Sub(from).Seconds())
}

func PrometheusMetrics(namespace string, labelsAndValues ...string) *Metrics {
	labels := []string{}
	for i := 0; i < len(labelsAndValues); i += 2 {
		labels = append(labels, labelsAndValues[i])
	}
	m := newMetrics(
		prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "logs_query_delay_seconds",
			Help:      "Histogram of how long it takes to receive queried Ethereum logs.",

			Buckets: stdprometheus.ExponentialBucketsRange(0.5, 30, 8),
		}, labels).With(labelsAndValues...),
		prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "logs_query_error_count",
			Help:      "Counter of how many errors were observed when querying Ethereum logs.",
		}, labels).With(labelsAndValues...),
	)

	m.setStartingValues()
	return m
}

func NoopMetrics() *Metrics {
	return newMetrics(
		discard.NewHistogram(),
		discard.NewCounter(),
	)
}
