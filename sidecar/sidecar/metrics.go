package sidecar

import (
	"math/big"
	"time"

	"github.com/go-kit/kit/metrics"
)

const (
	// MetricsSubsystem is a subsystem shared by all metrics exposed by this package.
	MetricsSubsystem = "core"
)

//go:generate go run ../../scripts/metricsgen -struct=Metrics

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
