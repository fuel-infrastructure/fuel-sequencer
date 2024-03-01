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
	ParamsKey = []byte("p_bridge")

	SupplyDeltaInfoKey = []byte("supply_delta_info")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
