package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName defines the module name
	ModuleName = "reports"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_reports"

	// SlashReportKey is the prefix to retrieve all SlashReport
	SlashReportKey = "SlashReport/value/"
)

var (
	ParamsKey = []byte("p_reports")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// SlashReportKeyPrefix returns the store key to retrieve a SlashReport using the height
func SlashReportKeyPrefix(slashReportHeight uint64) []byte {
	return append([]byte(SlashReportKey), sdk.Uint64ToBigEndian(slashReportHeight)...)
}
