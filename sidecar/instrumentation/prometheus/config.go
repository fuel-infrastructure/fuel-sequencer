package prometheus

import "time"

type Config struct {
	// Enabled enables serving of prometheus metrics under /metrics
	Enabled bool

	// ListenAddress is the address to listen for prometheus collectors
	ListenAddress string

	// MaxOpenConnections is the max number of simultaneous connections
	MaxOpenConnections int

	// ReadHeaderTimeout is the amount of time allowed to read request headers
	ReadHeaderTimeout time.Duration

	// Namespace is the instrumentation namespace.
	Namespace string
}
