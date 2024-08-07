package servers

import (
	"context"
	"net"

	"cosmossdk.io/errors"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

	// Default to insecure credentials if either pathToKeyFIle or pathToCertFile are empty, otherwise, use TLS.
	var s *grpc.Server
	if pathToKeyFile == "" || pathToCertFile == "" {
		s = grpc.NewServer()
	} else {
		creds, err := credentials.NewServerTLSFromFile(pathToCertFile, pathToKeyFile)
		if err != nil {
			panic(errors.Wrap(err, "failed to load TLS credentials"))
		}

		s = grpc.NewServer(grpc.Creds(creds))
	}

	// Register the TestSidecarServer as a SidecarServer in order to register all endpoints
	testSidecarServer := &TestSidecarServer{server: s, lis: lis}
	sidecartypes.RegisterSidecarServer(s, testSidecarServer)

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
