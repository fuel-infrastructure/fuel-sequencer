package ethwrappedclient

import (
	"time"

	"github.com/go-kit/kit/metrics"
)

const (
	// MetricsSubsystem is a subsystem shared by all metrics exposed by this package.
	MetricsSubsystem = "eth"
)

//go:generate go run ../../scripts/metricsgen -struct=Metrics

// Metrics contains metrics exposed by this package.
type Metrics struct {
	// How long it takes to receive queried Ethereum logs.
	LogsQueryDelaySeconds metrics.Histogram `metrics_buckettype:"exprange" metrics_bucketsizes:"0.5, 30, 8"`
	// How many errors were observed when querying Ethereum logs.
	LogsQueryErrorCount metrics.Counter
}

func (m *Metrics) setStartingValues() {}

func (m *Metrics) ObserveLogsQueryDelay(from, to time.Time) {
	m.LogsQueryDelaySeconds.Observe(to.Sub(from).Seconds())
}
