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
	IndexKey                    = []byte("index")
	SupplyDeltaInfoKey          = []byte("supply_delta_info")
	LastEthereumNonceKey        = []byte("LastEthereumNonce")
	LastEthereumBlockSyncedKey  = []byte("LastEthereumBlockSynced")
	EthereumEventIndexOffsetKey = []byte("EthereumEventIndexOffset")
	LastEthBlockUpdateTimeKey   = []byte("LastEthBlockUpdateTime")
	LastInjectedTxsSequenceKey  = []byte("LastInjectedTxsSequence")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
