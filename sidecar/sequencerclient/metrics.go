package sequencerclient

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/discard"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	// MetricsSubsystem is a subsystem shared by all metrics exposed by this package.
	MetricsSubsystem = "seq"
)

// Metrics contains metrics exposed by this package.
type Metrics struct {
	// Histogram of how long it takes to receive queries LastEthereumBlockSynced.
	LEBSQueryDelaySeconds metrics.Histogram
	// Counter of how many errors were observed when querying the LastEthereumBlockSynced.
	LEBSQueryErrorCount metrics.Counter
}

func (m *Metrics) setStartingValues() {}

func (m *Metrics) ObserveLEBSQueryDelay(from, to time.Time) {
	m.ObserveLEBSQueryDelaySeconds(to.Sub(from).Seconds())
}

func (m *Metrics) ObserveLEBSQueryDelaySeconds(seconds float64) {
	m.LEBSQueryDelaySeconds.Observe(seconds)
}

func PrometheusMetrics(namespace string, labelsAndValues ...string) *Metrics {
	labels := []string{}
	for i := 0; i < len(labelsAndValues); i += 2 {
		labels = append(labels, labelsAndValues[i])
	}
	m := &Metrics{
		LEBSQueryDelaySeconds: prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "lebs_query_delay_seconds",
			Help:      "Histogram of how long it takes to receive queries LastEthereumBlockSynced.",

			Buckets: stdprometheus.ExponentialBucketsRange(0.5, 30, 8),
		}, labels).With(labelsAndValues...),
		LEBSQueryErrorCount: prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "lebs_query_error_count",
			Help:      "Counter of how many errors were observed when querying the LastEthereumBlockSynced.",
		}, labels).With(labelsAndValues...),
	}

	m.setStartingValues()
	return m
}

func NopMetrics() *Metrics {
	m := &Metrics{
		LEBSQueryDelaySeconds: discard.NewHistogram(),
		LEBSQueryErrorCount:   discard.NewCounter(),
	}

	m.setStartingValues()
	return m
}
