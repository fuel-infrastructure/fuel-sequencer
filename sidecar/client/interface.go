package client

import (
	"context"

	"google.golang.org/grpc"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

// AppSidecarClient defines the interface that will be utilized by the
// application to query the Sidecar service. This interface is meant to
// be implemented by the gRPC client that connects to the Sidecar service.
type AppSidecarClient interface {
	sidecartypes.SidecarClient

	// Start starts the Sidecar client. This should connect to the remote
	// Sidecar service and return an error if the connection fails.
	Start(context.Context) error

	// Stop stops the Sidecar client.
	Stop() error
}

// NoOpClient is a no-op implementation of the AppSidecarClient interface. This
// implementation is used when the Sidecar service is disabled.
type NoOpClient struct{}

// Start is a no-op.
func (NoOpClient) Start(context.Context) error {
	return nil
}

// Stop is a no-op.
func (NoOpClient) Stop() error {
	return nil
}

// GetBlockEvents is a no-op.
func (c NoOpClient) GetBlockEvents(
	_ context.Context, _ *sidecartypes.QueryBlockEventsRequest, _ ...grpc.CallOption,
) (*sidecartypes.QueryBlockEventsResponse, error) {
	return &sidecartypes.QueryBlockEventsResponse{}, nil
}
