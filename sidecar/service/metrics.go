package service

import (
	"github.com/go-kit/kit/metrics"
)

const (
	// MetricsSubsystem is a subsystem shared by all metrics exposed by this package.
	MetricsSubsystem = "service"
)

//go:generate go run ../../scripts/metricsgen -struct=Metrics

// Metrics contains metrics exposed by this package.
type Metrics struct {
	// How many requests for block events the service received, with success status.
	BlockEventsRequestsTotal metrics.Counter `metrics_labels:"success"`
}

func (m *Metrics) setStartingValues() {}
