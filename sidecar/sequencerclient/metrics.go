package sequencerclient

import (
	"time"

	"github.com/go-kit/kit/metrics"
)

const (
	// MetricsSubsystem is a subsystem shared by all metrics exposed by this package.
	MetricsSubsystem = "seq"
)

//go:generate go run ../../scripts/metricsgen -struct=Metrics

// Metrics contains metrics exposed by this package.
type Metrics struct {
	// How long it takes to receive queries LastEthereumBlockSynced.
	LEBSQueryDelaySeconds metrics.Histogram `metrics_buckettype:"exprange" metrics_bucketsizes:"0.5, 30, 8"`
	// How many errors were observed when querying the LastEthereumBlockSynced.
	LEBSQueryErrorCount metrics.Counter
}

func (m *Metrics) setStartingValues() {}

func (m *Metrics) ObserveLEBSQueryDelay(from, to time.Time) {
	m.LEBSQueryDelaySeconds.Observe(to.Sub(from).Seconds())
}
