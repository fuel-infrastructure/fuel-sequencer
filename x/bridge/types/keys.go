package types

const (
	// ModuleName defines the module name
	ModuleName = "bridge"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_bridge"
)

var (
	ParamsKey                   = []byte("p_bridge")
	EthEventsTxIndexKey         = []byte("eth_events_tx_index")
	SupplyDeltaInfoKey          = []byte("supply_delta_info")
	LastEthereumNonceKey        = []byte("LastEthereumNonce")
	LastEthereumBlockSyncedKey  = []byte("LastEthereumBlockSynced")
	EthereumEventIndexOffsetKey = []byte("EthereumEventIndexOffset")
	SupplyDeltaProcessedKey     = []byte("SupplyDeltaProcessed")
	LastEthBlockUpdateTimeKey   = []byte("LastEthBlockUpdateTime")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
