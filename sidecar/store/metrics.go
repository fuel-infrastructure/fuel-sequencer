package store

import (
	"math/big"

	"github.com/go-kit/kit/metrics"
)

const (
	// MetricsSubsystem is a subsystem shared by all metrics exposed by this package.
	MetricsSubsystem = "store"
)

//go:generate go run ../../scripts/metricsgen -struct=Metrics

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

func (m *Metrics) SetLastSyncedBlock(block *big.Int) {
	blockF64, _ := block.Float64()
	m.LastSyncedBlock.Set(blockF64)
}

func (m *Metrics) SetStartQueryBlock(block *big.Int) {
	blockF64, _ := block.Float64()
	m.StartQueryBlock.Set(blockF64)
}
