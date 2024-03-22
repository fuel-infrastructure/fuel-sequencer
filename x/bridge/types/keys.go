package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName defines the module name
	ModuleName = "bridge"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_bridge"
)

var (
	ParamsKey                  = []byte("p_bridge")
	EthEventsTxPrefixKey       = []byte("eth_events_tx")
	SupplyDeltaInfoKey         = []byte("supply_delta_info")
	LastEthereumNonceKey       = []byte("LastEthereumNonce")
	LastEthereumBlockSyncedKey = []byte("LastEthereumBlockSynced")
	SupplyDeltaProcessedKey    = []byte("SupplyDeltaProcessed")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// EthEventsTxKey the store key for a EthEventsTx by ethereum block height.
func EthEventsTxKey(blockHeight uint64) []byte {
	return append(EthEventsTxPrefixKey, sdk.Uint64ToBigEndian(blockHeight)...)
}
