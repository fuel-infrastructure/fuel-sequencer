package cmd

import "time"

const (
	// Sidecar flags

	FlagSidecarHost                  = "host"
	FlagSidecarPort                  = "port"
	FlagSidecarDevelopment           = "development"
	FlagEthereumWebSocketUrl         = "eth_ws_url"
	FlagEthereumContractAddr         = "eth_contract_address"
	FlagEthereumMaxBlockRange        = "eth_max_block_range"
	FlagEthereumMinLogsQueryInterval = "eth_min_logs_query_interval"
	FlagEthereumUnsafeStartBlock     = "unsafe_eth_start_block"
	FlagSequencerGrpcUrl             = "sequencer_grpc_url"
	FlagSequencerRpcUrl              = "sequencer_rpc_url"

	// Sidecar client flags

	FlagSidecarGrpcUrl = "sidecar_grpc_url"
)

type sidecarConfig struct {
	host        string
	port        string
	development bool
}

type sequencerConfig struct {
	grpcUrl string
	rpcUrl  string
}

type ethereumConfig struct {
	webSocketUrl         string
	contractAddrHex      string
	maxBlockRange        int64
	minLogsQueryInterval time.Duration
	unsafeStartBlock     int64
}
