package client

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"cosmossdk.io/log"
	sidecarconfig "github.com/fuel-infrastructure/fuel-sequencer/sidecar/config"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils/credentials"
	"google.golang.org/grpc"
)

var _ AppSidecarClient = (*GRPCClient)(nil)

// GRPCClient defines an implementation of a gRPC Sidecar client. This client can
// be used in ABCI++ calls where the application wants the Sidecar process to be
// run out-of-process. The client must be started upon app construction and
// stopped upon app shutdown/cleanup.
type GRPCClient struct {
	logger log.Logger
	mutex  sync.Mutex

	// address of remote Sidecar server
	addr string
	// underlying Sidecar client
	client sidecartypes.SidecarClient
	// underlying grpc connection
	conn *grpc.ClientConn
	// timeout for the client, event requests will block for this duration.
	timeout time.Duration
	// pathToCertFile defines the path to the certificate file of the sidecar server.
	pathToCertFile string
}

// NewClientFromConfig creates a new grpc client of the Sidecar service with the given
// app configuration. This returns an error if the configuration is invalid.
func NewClientFromConfig(
	cfg sidecarconfig.SidecarConfig,
	logger log.Logger,
) (AppSidecarClient, error) {
	if err := cfg.ValidateBasic(); err != nil {
		return nil, err
	}

	if !cfg.Enabled {
		logger.Warn("sidecar not enabled in app.toml")
		return &NoOpClient{}, nil
	}

	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	return NewClient(logger, cfg.Address, cfg.Timeout, cfg.PathToCertFile)
}

// NewClient creates a new grpc client of the Sidecar
// service with the given address and timeout.
func NewClient(
	logger log.Logger,
	address string,
	timeout time.Duration,
	pathToCertFile string,
) (AppSidecarClient, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// If the sidecar was enabled, the server address cannot be empty
	if address == "" {
		return nil, fmt.Errorf("sidecar address cannot be empty")
	}

	// Prepending a "//" allows addresses without a scheme.
	if _, err := url.Parse("//" + address); err != nil {
		return nil, fmt.Errorf("invalid Sidecar address: %w", err)
	}

	if timeout <= 0 {
		return nil, fmt.Errorf("timeout must be positive")
	}

	client := &GRPCClient{
		logger:         logger,
		addr:           address,
		timeout:        timeout,
		pathToCertFile: pathToCertFile,
	}

	return client, nil
}

// Start starts the GRPC client. This method dials the remote
// Sidecar service and errors if the connection fails.
func (c *GRPCClient) Start(ctx context.Context) error {
	c.logger.Info("starting GRPC Sidecar client", "sidecar server address", c.addr)

	// Set up a secure connection with the sidecar server if configured by the operator
	sidecarConnCreds, _, err := credentials.NewClientTransportCredentialsFromCertFile(c.pathToCertFile)
	if err != nil {
		return fmt.Errorf("failed to get sidecar server TLS credentials; error: %w", err)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(sidecarConnCreds),
	}

	// dial the client, but defer to context closure, if necessary
	var (
		conn *grpc.ClientConn
		done = make(chan struct{})
	)
	go func() {
		defer close(done)

		//nolint:staticcheck
		conn, err = grpc.DialContext(ctx, c.addr, opts...)
	}()

	// wait for either the context to close or the dial to complete
	select {
	case <-ctx.Done():
		err = fmt.Errorf("context closed before Sidecar client could start: %w", ctx.Err())
	case <-done:
	}
	if err != nil {
		c.logger.Error("failed to dial Sidecar gRPC server", "err", err)
		return fmt.Errorf("failed to dial Sidecar gRPC server: %w", err)
	}

	c.mutex.Lock()
	c.client = sidecartypes.NewSidecarClient(conn)
	c.conn = conn
	c.mutex.Unlock()

	c.logger.Info("GRPC Sidecar client started", "addr", c.addr)

	return nil
}

// Stop stops the GRPC client. This method closes the connection to the remote.
func (c *GRPCClient) Stop() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.logger.Info("stopping Sidecar client")
	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.logger.Info("sidecar client stopped", "err", err)

	return err
}

// GetBlockEvents returns the block events from the remote Sidecar service.
// This method blocks for the timeout duration configured on the client,
// otherwise it returns the response from the remote Sidecar.
func (c *GRPCClient) GetBlockEvents(
	ctx context.Context,
	req *sidecartypes.QueryBlockEventsRequest,
	_ ...grpc.CallOption,
) (resp *sidecartypes.QueryBlockEventsResponse, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// set deadline on the context
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if c.client == nil {
		return nil, fmt.Errorf("sidecar client not started")
	}

	return c.client.GetBlockEvents(ctx, req, grpc.WaitForReady(true))
}
