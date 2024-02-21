package types

const (
	// ModuleName defines the module name
	ModuleName = "sequencing"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_sequencing"
)

var (
	ParamsKey = []byte("p_sequencing")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
