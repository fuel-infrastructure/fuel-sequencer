package testutil

import (
	"net"

	"cosmossdk.io/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// TestSidecarServer is a lightweight gRPC server used to simulate the Sidecar server. This approach is sometimes
// preferred over traditional mocking to ensure that functionality works as intended by running an actual server.
type TestSidecarServer struct {
	server *grpc.Server
	lis    net.Listener
}

func MustMakeTestTLSSidecarServer(address string, pathToCertFile, pathToKeyFile string) *TestSidecarServer {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		panic(errors.Wrapf(err, "failed to listen on %s", address))
	}

	// Create a secure grpc.Server
	creds, err := credentials.NewServerTLSFromFile(pathToCertFile, pathToKeyFile)
	if err != nil {
		panic(errors.Wrap(err, "failed to load TLS credentials"))
	}

	s := grpc.NewServer(grpc.Creds(creds))

	return &TestSidecarServer{server: s, lis: lis}
}

func MustMakeTestSidecarServer(address string) *TestSidecarServer {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		panic(errors.Wrapf(err, "failed to listen on %s", address))
	}

	// Create an insecure grpc.Server
	s := grpc.NewServer()

	return &TestSidecarServer{server: s, lis: lis}
}

func (d *TestSidecarServer) Start() {
	go func() {
		if err := d.server.Serve(d.lis); err != nil {
			panic(errors.Wrap(err, "failed to serve"))
		}
	}()
}

func (d *TestSidecarServer) Stop() {
	d.server.Stop()
}

func (d *TestSidecarServer) GetAddress() string {
	return d.lis.Addr().String()
}
