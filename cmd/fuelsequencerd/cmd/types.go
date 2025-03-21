package cmd

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
)

const (
	// Sidecar flags

	FlagSidecarHost                  = "host"
	FlagSidecarPort                  = "port"
	FlagSidecarDevelopment           = "development"
	FlagSidecarPathToKeyFile         = "sidecar_path_to_key_file"
	FlagSidecarPathToCertFile        = "sidecar_path_to_cert_file"
	FlagEthereumWebSocketUrl         = "eth_ws_url"
	FlagEthereumRpcUrl               = "eth_rpc_url"
	FlagEthereumContractAddr         = "eth_contract_address"
	FlagEthereumMaxBlockRange        = "eth_max_block_range"
	FlagEthereumMinLogsQueryInterval = "eth_min_logs_query_interval"
	FlagEthereumRpcQueryTimeout      = "eth_rpc_query_timeout"
	FlagEthereumUnsafeStartBlock     = "unsafe_eth_start_block"
	FlagEthereumUnsafeEndBlock       = "unsafe_eth_end_block"
	FlagSequencerGrpcUrl             = "sequencer_grpc_url"
	FlagSequencerPathToCertFile      = "sequencer_path_to_cert_file"
	FlagSequencerUnsafeBridgeDenom   = "unsafe_sequencer_bridge_denom"
	FlagPrometheusEnabled            = "prometheus_enabled"
	FlagPrometheusListenAddress      = "prometheus_listen_address"
	FlagPrometheusMaxOpenConnections = "prometheus_max_open_connections"
	FlagPrometheusReadHeaderTimeout  = "prometheus_read_header_timeout"
	FlagPrometheusWriteTimeout       = "prometheus_write_timeout"
	FlagPrometheusNamespace          = "prometheus_namespace"

	// Sidecar client flags

	FlagSidecarGrpcUrl              = "sidecar_grpc_url"
	FlagQueryTimeout                = "query_timeout"
	FlagSidecarClientPathToCertFile = "sidecar_path_to_cert_file"
)

type sidecarConfig struct {
	host           string
	port           string
	development    bool
	pathToCertFile string
	pathToKeyFile  string
}

func (cfg *sidecarConfig) Validate() error {
	// Validate host
	if cfg.host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	// First try parsing as a URL
	if _, err := url.ParseRequestURI(cfg.host + ":" + cfg.port); err != nil {
		// Then try to validate as an IP address
		if ip := net.ParseIP(cfg.host); ip == nil {
			return fmt.Errorf("invalid host format: must be a valid URL or IP address")
		}
	}

	// Validate port
	if err := validatePort(cfg.port); err != nil {
		return err
	}

	// Validate certificate file
	if cfg.pathToCertFile != "" {
		if _, err := os.Stat(cfg.pathToCertFile); err != nil {
			return fmt.Errorf("certificate file not found at %s: %w", cfg.pathToCertFile, err)
		}
	}

	// Validate key file
	if cfg.pathToKeyFile != "" {
		if _, err := os.Stat(cfg.pathToKeyFile); err != nil {
			return fmt.Errorf("key file not found at %s: %w", cfg.pathToKeyFile, err)
		}
	}

	return nil
}

type sequencerConfig struct {
	grpcUrl           string
	pathToCertFile    string
	unsafeBridgeDenom string
}

func (cfg *sequencerConfig) Validate() error {
	// Validate gRPC URL
	if cfg.grpcUrl == "" {
		return fmt.Errorf("gRPC URL cannot be empty")
	}

	host, port := "", ""
	if strings.Contains(cfg.grpcUrl, "://") {
		// If it's a URL with protocol, parse it
		parsedURL, err := url.Parse(cfg.grpcUrl)
		if err != nil {
			return fmt.Errorf("invalid gRPC URL format: must be a valid URL or host:port format")
		}
		host = parsedURL.Hostname()
		port = parsedURL.Port()
	} else {
		// If it's a host:port format, split it
		h, p, err := net.SplitHostPort(cfg.grpcUrl)
		if err != nil {
			return fmt.Errorf("invalid gRPC URL format: must be a valid URL or host:port format")
		}
		host = h
		port = p
	}

	if err := validatePort(port); err != nil {
		return err
	}

	// Try parsing as IP (handles both IPv4 and IPv6)
	if ip := net.ParseIP(host); ip == nil && host != "localhost" {
		// Not an IP and not localhost, validate as hostname
		return fmt.Errorf("invalid hostname in gRPC URL")
	}

	// Validate certificate file if provided
	if cfg.pathToCertFile != "" {
		if _, err := os.Stat(cfg.pathToCertFile); err != nil {
			return fmt.Errorf("certificate file not found at %s: %w", cfg.pathToCertFile, err)
		}
	}

	// Validate bridge denom
	if err := types.ValidateDenom(cfg.unsafeBridgeDenom); err != nil {
		return fmt.Errorf("invalid bridge denom format: %w", err)
	}

	return nil
}

type ethereumConfig struct {
	webSocketUrl         string
	rpcUrl               string
	contractAddrHex      string
	maxBlockRange        int64
	minLogsQueryInterval time.Duration
	rpcQueryTimeout      time.Duration
	unsafeStartBlock     int64
	unsafeEndBlock       int64
}

func (cfg *ethereumConfig) Validate() error {
	if cfg.unsafeStartBlock < 0 {
		return fmt.Errorf("ethereum unsafe start block must be >= 0, got: %d", cfg.unsafeStartBlock)
	}
	if cfg.unsafeEndBlock < 0 {
		return fmt.Errorf("ethereum unsafe end block must be >= 0, got: %d", cfg.unsafeEndBlock)
	}
	if cfg.maxBlockRange < 1 {
		return fmt.Errorf("ethereum max block range must be >= 1, got: %d", cfg.maxBlockRange)
	}
	if cfg.rpcQueryTimeout <= 0 {
		return fmt.Errorf("ethereum rpc query timeout must be > 0, got: %s", cfg.rpcQueryTimeout.String())
	}
	return nil
}

// AppOptionsMap is a stub implementing AppOptions which can get data from a map.
// It is used to inject app options very early on, for the NewRootCmd function.
type AppOptionsMap map[string]interface{}

func (m AppOptionsMap) Get(key string) interface{} {
	v, ok := m[key]
	if !ok {
		return interface{}(nil)
	}

	return v
}
