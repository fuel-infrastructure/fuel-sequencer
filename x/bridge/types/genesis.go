package types

import (
	"cosmossdk.io/math"
)

// this line is used by starport scaffolding # genesis/types/import

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
		SupplyDeltaInfo: &SupplyDeltaInfo{
			LastSupply: math.ZeroInt(),
			Delta:      math.ZeroInt(),
			Offset:     math.ZeroInt(),
		},
		LastEthereumNonce:        math.ZeroInt(),
		LastEthereumBlockSynced:  0,
		EthereumEventIndexOffset: 0,
		// this line is used by starport scaffolding # genesis/types/default
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
