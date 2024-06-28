package types

import (
	"fmt"
	"time"

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
		LastEthBlockUpdateTime:   time.Time{},
		// this line is used by starport scaffolding # genesis/types/default
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {

	// Validate supply_delta_info.
	if err := gs.SupplyDeltaInfo.ValidateBasic(); err != nil {
		return err
	}

	// Validate last_ethereum_nonce.
	if err := ValidateLastEthereumNonce(gs.LastEthereumNonce); err != nil {
		return err
	}

	// Validate last_ethereum_block_synced.
	if err := ValidateLastEthereumBlockSynced(gs.LastEthereumBlockSynced); err != nil {
		return err
	}

	// Validate ethereum_event_index_offset.
	if err := ValidateEthereumEventIndexOffset(gs.EthereumEventIndexOffset); err != nil {
		return err
	}

	// Validate last_eth_block_update_time.
	if err := ValidateLastEthBlockUpdateTime(gs.LastEthBlockUpdateTime); err != nil {
		return err
	}

	// this line is used by starport scaffolding # genesis/types/validate
	return gs.Params.Validate()
}

func ValidateLastEthereumNonce(i interface{}) error {
	_, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func ValidateLastEthereumBlockSynced(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func ValidateEthereumEventIndexOffset(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func ValidateLastEthBlockUpdateTime(i interface{}) error {
	_, ok := i.(time.Time)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
