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
			Offset:     math.ZeroInt(),
			ToReport:   math.ZeroInt(),
		},
		LastEthereumNonce:        math.ZeroInt(),
		LastEthereumBlockSynced:  0,
		EthereumEventIndexOffset: 0,
		LastEthBlockUpdateTime:   time.Time{},
		LastConsensusTxsSequence: 0,
		// this line is used by starport scaffolding # genesis/types/default
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// this line is used by starport scaffolding # genesis/types/validate

	if err := gs.SupplyDeltaInfo.ValidateBasic(); err != nil {
		return err
	}

	if err := ValidateLastEthereumNonce(gs.LastEthereumNonce); err != nil {
		return err
	}

	if err := ValidateLastEthereumBlockSynced(gs.LastEthereumBlockSynced); err != nil {
		return err
	}

	if err := ValidateEthereumEventIndexOffset(gs.EthereumEventIndexOffset); err != nil {
		return err
	}

	if err := ValidateLastEthBlockUpdateTime(gs.LastEthBlockUpdateTime); err != nil {
		return err
	}

	if err := ValidateLastConsensusTxsSequence(gs.LastConsensusTxsSequence); err != nil {
		return err
	}

	return gs.Params.Validate()
}

// ValidateLastEthereumNonce validates that the last Ethereum nonce is non-negative.
// The expected nonce should be 0 since a (+1) is always added to the nonce before it is used.
func ValidateLastEthereumNonce(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("expected LastEthereumNonce >= 0, received %s", v.String())
	}

	return nil
}

// ValidateLastEthereumBlockSynced validates that the type of LastEthereumBlockSynced is correct.
// Since we always sync to the next Ethereum block (+1), LastEthereumBlockSynced can be 0.
func ValidateLastEthereumBlockSynced(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

// ValidateEthereumEventIndexOffset validates that the type of ValidateEthereumEventIndexOffset is correct.
func ValidateEthereumEventIndexOffset(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

// ValidateLastEthBlockUpdateTime validates that the type of ValidateLastEthBlockUpdateTime is correct.
func ValidateLastEthBlockUpdateTime(i interface{}) error {
	_, ok := i.(time.Time)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

// ValidateLastConsensusTxsSequence validates that the last consensus txs sequence is uint64.
func ValidateLastConsensusTxsSequence(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
