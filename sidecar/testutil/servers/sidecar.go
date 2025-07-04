package servers

import (
	"context"
	"fmt"
	"net"

	"cosmossdk.io/errors"
	"google.golang.org/grpc"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils/credentials"
)

// TestSidecarServer is a lightweight gRPC server used to simulate the Sidecar server. This approach is sometimes
// preferred over traditional mocking to ensure that functionality works as intended by running an actual server.
type TestSidecarServer struct {
	server *grpc.Server
	lis    net.Listener
	events []*sidecartypes.Event
}

// MustMakeTestSidecarServer creates an insecure GRPC server if pathToCertFile or pathToKeyFile are an empty string,
// otherwise, it creates a TLS server.
func MustMakeTestSidecarServer(address string, pathToCertFile, pathToKeyFile string) *TestSidecarServer {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		panic(errors.Wrapf(err, "failed to listen on %s", address))
	}

	// Make use of certificates if indicated by the operator
	serverCredentials, _, _, err := credentials.NewServerTransportCredentialsFromCertFile(
		pathToCertFile, pathToKeyFile,
	)
	if err != nil {
		panic(fmt.Errorf("failed to get sidecar server TLS credentials; error: %w", err))
	}

	server := grpc.NewServer(grpc.Creds(serverCredentials))

	// Register the TestSidecarServer as a SidecarServer in order to register all endpoints
	testSidecarServer := &TestSidecarServer{server: server, lis: lis}
	sidecartypes.RegisterSidecarServer(server, testSidecarServer)

	return testSidecarServer
}

func (d *TestSidecarServer) Start() {
	go func() {
		// Serve. If the server has stopped we don't want to panic as this means that the test stopped the server.
		if err := d.server.Serve(d.lis); err != nil && err != grpc.ErrServerStopped {
			panic(errors.Wrap(err, "failed to serve"))
		}
	}()
}

func (d *TestSidecarServer) Stop() {
	d.server.GracefulStop()
}

func (d *TestSidecarServer) GetAddress() string {
	return d.lis.Addr().String()
}

// GetBlockEvents is used as a mock to the Sidecar server's GetBlockEvents RPC method. It returns the value stored
// in testSidecarServer.events
func (d *TestSidecarServer) GetBlockEvents(_ context.Context, _ *sidecartypes.QueryBlockEventsRequest) (
	*sidecartypes.QueryBlockEventsResponse, error,
) {
	return &sidecartypes.QueryBlockEventsResponse{Events: d.events}, nil
}

// SetEvents sets d.events to the specified value. It is useful for mocking the GetBlockEvents RPC call.
func (d *TestSidecarServer) SetEvents(events []*sidecartypes.Event) {
	d.events = events
}
