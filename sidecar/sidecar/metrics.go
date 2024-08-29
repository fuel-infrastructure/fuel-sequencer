package sidecar

import (
	"math/big"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/discard"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	// MetricsSubsystem is a subsystem shared by all metrics exposed by this package.
	MetricsSubsystem = "core"
)

// Metrics contains metrics exposed by this package.
type Metrics struct {
	// Whether the sidecar is catching up to Ethereum. 1 if yes, 0 if no.
	CatchingUp metrics.Gauge
	// The last Ethereum block that the Sidecar can sync at the moment.
	MaxSyncableBlock metrics.Gauge
	// The most recent Ethereum header that the Sidecar detected.
	LastHeaderSeen metrics.Gauge
	// Delays in receiving Ethereum headers.
	HeaderDelaySeconds metrics.Histogram
	// The time that the Sidecar started as a Unix timestamp in seconds.
	startTime metrics.Gauge
}

func newMetrics(
	catchingUp metrics.Gauge,
	maxSyncableBlock metrics.Gauge,
	lastHeaderSeen metrics.Gauge,
	headerDelaySeconds metrics.Histogram,
	startTime metrics.Gauge,
) *Metrics {
	return &Metrics{
		CatchingUp:         catchingUp,
		MaxSyncableBlock:   maxSyncableBlock,
		LastHeaderSeen:     lastHeaderSeen,
		HeaderDelaySeconds: headerDelaySeconds,
		startTime:          startTime,
	}
}

func (m *Metrics) setStartingValues() {
	m.startTime.Set(float64(time.Now().Unix()))
}

func (m *Metrics) ObserveHeaderDelay(from, to time.Time) {
	m.HeaderDelaySeconds.Observe(to.Sub(from).Seconds())
}

func (m *Metrics) SetLastHeaderSeen(block *big.Int) {
	blockF64, _ := block.Float64()
	m.LastHeaderSeen.Set(blockF64)
}

func (m *Metrics) SetMaxSyncableBlock(block *big.Int) {
	blockF64, _ := block.Float64()
	m.MaxSyncableBlock.Set(blockF64)
}

func PrometheusMetrics(namespace string, labelsAndValues ...string) *Metrics {
	labels := []string{}
	for i := 0; i < len(labelsAndValues); i += 2 {
		labels = append(labels, labelsAndValues[i])
	}
	m := newMetrics(
		prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "catching_up",
			Help:      "Whether the sidecar is catching up to ethereum. 1 if yes, 0 if no.",
		}, labels).With(labelsAndValues...),
		prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "max_syncable_block",
			Help:      "The last Ethereum block that the Sidecar can sync at the moment.",
		}, labels).With(labelsAndValues...),
		prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "last_header_seen",
			Help:      "The most recent Ethereum header that the Sidecar detected.",
		}, labels).With(labelsAndValues...),
		prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "header_delay_seconds",
			Help:      "Delays in receiving Ethereum headers.",

			Buckets: stdprometheus.ExponentialBucketsRange(0.5, 30, 8),
		}, labels).With(labelsAndValues...),
		prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "start_time",
			Help:      "The time that the Sidecar started as a Unix timestamp in seconds.",
		}, labels).With(labelsAndValues...),
	)

	m.setStartingValues()
	return m
}

func NoopMetrics() *Metrics {
	return newMetrics(
		discard.NewGauge(),
		discard.NewGauge(),
		discard.NewGauge(),
		discard.NewHistogram(),
		discard.NewGauge(),
	)
}
