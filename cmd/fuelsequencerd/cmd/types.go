package cmd

import "time"

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
	FlagEthereumUnsafeStartBlock     = "unsafe_eth_start_block"
	FlagEthereumUnsafeEndBlock       = "unsafe_eth_end_block"
	FlagSequencerGrpcUrl             = "sequencer_grpc_url"
	FlagSequencerPathToCertFile      = "sequencer_path_to_cert_file"
	FlagPrometheusEnabled            = "prometheus_enabled"
	FlagPrometheusListenAddress      = "prometheus_listen_address"
	FlagPrometheusMaxOpenConnections = "prometheus_max_open_connections"
	FlagPrometheusReadHeaderTimeout  = "prometheus_read_header_timeout"
	FlagPrometheusNamespace          = "prometheus_namespace"

	// Sidecar client flags

	FlagSidecarGrpcUrl              = "sidecar_grpc_url"
	FlagQueryTimeout                = "query_timeout"
	FlagSidecarClientPathToCertFile = "sidecar_path_to_cert_file"

	// Testing flags - Flags in this category should never be merged into main

	FlagTestNoOfMsgSends = "test_no_of_msg_sends"
)

type sidecarConfig struct {
	host           string
	port           string
	development    bool
	pathToCertFile string
	pathToKeyFile  string
}

type sequencerConfig struct {
	grpcUrl        string
	pathToCertFile string
}

type ethereumConfig struct {
	webSocketUrl         string
	rpcUrl               string
	contractAddrHex      string
	maxBlockRange        int64
	minLogsQueryInterval time.Duration
	unsafeStartBlock     int64
	unsafeEndBlock       int64
	testNoOfMsgSends     int64 // This value should only be used for testing purposes and never be merged into main
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
