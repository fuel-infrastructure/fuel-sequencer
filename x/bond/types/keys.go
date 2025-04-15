package types

const (
	// ModuleName defines the module name
	ModuleName = "bond"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_bond"
)

var (
	ParamsKey    = []byte("p_bond")
	InflationKey = []byte("Inflation")
	AuthorityKey = []byte("Authority")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
