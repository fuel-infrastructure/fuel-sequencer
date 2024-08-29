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
	LastEthereumHeaderSeen metrics.Gauge
	// Histogram of delays in receiving Ethereum headers.
	HeaderReceiveDelaySeconds metrics.Histogram
}

func (m *Metrics) setStartingValues() {}

func (m *Metrics) ObserveHeaderReceiveDelay(from, to time.Time) {
	m.ObserveHeaderReceiveDelaySeconds(to.Sub(from).Seconds())
}

func (m *Metrics) ObserveHeaderReceiveDelaySeconds(seconds float64) {
	m.HeaderReceiveDelaySeconds.Observe(seconds)
}

func (m *Metrics) SetLastEthereumHeaderSeen(block *big.Int) {
	blockF64, _ := block.Float64()
	m.LastEthereumHeaderSeen.Set(blockF64)
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
		LastEthereumHeaderSeen: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "last_ethereum_header_seen",
			Help:      "The most recent Ethereum header that the Sidecar detected.",
		}, labels).With(labelsAndValues...),
		HeaderReceiveDelaySeconds: prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "header_receive_delay_seconds",
			Help:      "Histogram of delays in receiving Ethereum headers.",

			Buckets: stdprometheus.ExponentialBucketsRange(0.5, 30, 8),
		}, labels).With(labelsAndValues...),
	}

	m.setStartingValues()
	return m
}

func NopMetrics() *Metrics {
	m := &Metrics{
		CatchingUp:                discard.NewGauge(),
		MaxSyncableBlock:          discard.NewGauge(),
		LastEthereumHeaderSeen:    discard.NewGauge(),
		HeaderReceiveDelaySeconds: discard.NewHistogram(),
	}

	m.setStartingValues()
	return m
}
