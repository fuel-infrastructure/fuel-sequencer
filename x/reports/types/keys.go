package types

const (
	// ModuleName defines the module name
	ModuleName = "reports"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_reports"
)

var (
	ParamsKey = []byte("p_reports")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
