package types

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

// SlashEntryKeyPrefix returns the store key to retrieve a SlashEntry using the height, delegator address and validator
// address. Note that []byte{} is used as a separator between keys so that the key is as compact and lean as possible.
// This is only possible because heights and addresses are of predictable length i.e. the key's components can be
// reconstructed.
func SlashEntryKeyPrefix(slashReportHeight uint64, delegatorAddress, validatorAddress string) []byte {
	return bytes.Join([][]byte{
		[]byte(SlashReportKey),
		sdk.Uint64ToBigEndian(slashReportHeight),
		[]byte(delegatorAddress),
		[]byte(validatorAddress),
	}, []byte{})
}

// ExtractHeightFromSlashEntryKey extracts height from a slash entry key.
// Given the height is uint64 it should occupy 8 bytes.
func ExtractHeightFromSlashEntryKey(key []byte) uint64 {
	prefixLen := len(SlashReportKey)
	heightBytes := key[prefixLen : prefixLen+8]
	return sdk.BigEndianToUint64(heightBytes)
}
