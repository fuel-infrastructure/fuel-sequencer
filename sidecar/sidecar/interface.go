package sidecar

import (
	"context"
	"math/big"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

var _ SidecarService = (*Sidecar)(nil)

// SidecarService defines the expected interface for the sidecar.
type SidecarService interface {
	// IsStopped checks if the sidecar has been stopped.
	IsStopped() bool

	// QueryBlockEvents queries the `blocksMap` for events associated with a specific block number.
	QueryBlockEvents(ctx context.Context, blockNumber *big.Int) ([]sidecartypes.Event, error)

	// StartFetching begins the process of querying and storing events from the Ethereum blockchain.
	StartFetching(ctx context.Context) error

	// ShutDown signals the sidecar to stop.
	ShutDown()
}
