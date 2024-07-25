package client

import (
	"time"

	"cosmossdk.io/log"
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
