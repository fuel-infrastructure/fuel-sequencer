package client

import (
	"sync"
	"time"

	"cosmossdk.io/log"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"google.golang.org/grpc"
)

// Addr is an export of c.addr. This should only be used for testing purposes.
func (c *GRPCClient) Addr() string {
	return c.addr
}

// Logger is an export of c.logger. This should only be used for testing purposes.
func (c *GRPCClient) Logger() log.Logger {
	return c.logger
}

// PathToCertFile is an export of c.pathToCertFile. This should only be used for testing purposes.
func (c *GRPCClient) PathToCertFile() string {
	return c.pathToCertFile
}

// Timeout is an export of c.timeout. This should only be used for testing purposes.
func (c *GRPCClient) Timeout() time.Duration {
	return c.timeout
}

// Conn is an export of c.conn. This should only be used for testing purposes.
func (c *GRPCClient) Conn() *grpc.ClientConn {
	return c.conn
}

// Client is an export of c.client. This should only be used for testing purposes.
func (c *GRPCClient) Client() sidecartypes.SidecarClient {
	return c.client
}

// Mutex is an export of c.mutex. This should only be used for testing purposes.
func (c *GRPCClient) Mutex() sync.Mutex {
	return c.mutex
}
