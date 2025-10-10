package sequencer

import (
	cmconfig "github.com/cometbft/cometbft/config"
	sequencingtypes "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

// As noted for load testing
// https://github.com/fuel-infrastructure/notes/blob/main/testing/LoadTests.md#blob-posting

func loadTestSequencerConfigs(sequencingGenState *sequencingtypes.GenesisState) {
	// Sequencer Configs
	// sequencing.sequencer_tx_max_bytes: 9961472 (9.5MiB)
	sequencingGenState.Params.SequencerTxMaxBytes = uint64(9961472)
	// sequencing.max_blob_size_bytes: 9437184 (9MiB)
	sequencingGenState.Params.MaxBlobSizeBytes = uint64(9437184)
}

func loadTestConsensusConfigs(valConfig *cmconfig.Config) {
	// TODO: implement - currently performed manually
	// Consensus Configs
	// consensus.block.max_bytes: 10485760 (10MiB)
}

func loadTestCometBFTConfigs(valConfig *cmconfig.Config) {
	// CometBFT Configs
	// config.toml -> [rpc] -> max_body_bytes: 20000000 (20MB)
	valConfig.RPC.MaxBodyBytes = int64(20000000)
	// config.toml -> [mempool] -> max_tx_bytes: 9961472 (9.5MiB)
	valConfig.Mempool.MaxTxBytes = int(9961472)
	// config.toml -> [mempool] -> max_txs_bytes: 104857600 (100MiB)
	valConfig.Mempool.MaxTxsBytes = int64(104857600)
}
