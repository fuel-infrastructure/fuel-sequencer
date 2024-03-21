package types

import "cosmossdk.io/math"

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
		LastEthereumNonce:       math.ZeroInt(),
		LastEthereumBlockSynced: math.ZeroInt(),
		EthEventsTx:             nil,
		// this line is used by starport scaffolding # genesis/types/default
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// this line is used by starport scaffolding # genesis/types/validate

	// If there is an EthEventsTx found then verify that the number it belongs to is
	// lastEthereumBlockSynced + 1
	if gs.EthEventsTx != nil && !gs.EthEventsTx.BlockNumber.Equal(gs.LastEthereumBlockSynced.Add(math.OneInt())) {
		return ErrEthEventsTxBlockNotSequential.Wrapf("expected %s but found %s", gs.EthEventsTx, gs.LastEthereumBlockSynced.Add(math.OneInt()))
	}

	return gs.Params.Validate()
}
