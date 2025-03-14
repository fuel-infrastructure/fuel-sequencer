package types_test

import (
	"testing"
	"time"

	testutiltypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"

	"cosmossdk.io/math"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"

	"github.com/stretchr/testify/require"
)

func TestValidateGenesisState(t *testing.T) {
	tests := []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		// Default genesis is not valid since some default parameters are invalid
		{
			desc:     "default is not valid",
			genState: types.DefaultGenesis(),
			valid:    false,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				Params: types.NewParams(
					types.DefaultBridgeDenom,
					testutiltypes.TestBridgeDenomTotalSupply,
					types.DefaultEthereumProxyContractAddress,
					types.DefaultSupplyDeltaPeriod,
					testutiltypes.TestVestingStartingTime,
					nil,
					types.DefaultMaxEthBlockUpdateDelay,
					types.DefaultSequencerTxsAllocation,
				),
				SupplyDeltaInfo: &types.SupplyDeltaInfo{
					LastSupply: math.NewInt(99),
					Offset:     math.NewInt(123),
					ToReport:   math.NewInt(34),
				},
				LastEthereumNonce:        math.NewInt(3),
				LastEthereumBlockSynced:  1,
				EthereumEventIndexOffset: 2,
				LastEthBlockUpdateTime:   time.Now(),
				LastConsensusTxsSequence: 5,
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "invalid genesis state - negative last supply",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				SupplyDeltaInfo: &types.SupplyDeltaInfo{
					LastSupply: math.NewInt(-123),
					Offset:     math.NewInt(123),
					ToReport:   math.NewInt(34),
				},
				LastEthereumNonce:        math.NewInt(3),
				LastEthereumBlockSynced:  1,
				EthereumEventIndexOffset: 2,
				LastEthBlockUpdateTime:   time.Now(),
				LastConsensusTxsSequence: 5,
			},
			valid: false,
		},
		{
			desc: "invalid genesis state - negative last ethereum nonce",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				SupplyDeltaInfo: &types.SupplyDeltaInfo{
					LastSupply: math.NewInt(99),
					Offset:     math.NewInt(123),
					ToReport:   math.NewInt(34),
				},
				LastEthereumNonce:        math.NewInt(-1),
				LastEthereumBlockSynced:  1,
				EthereumEventIndexOffset: 2,
				LastEthBlockUpdateTime:   time.Now(),
				LastConsensusTxsSequence: 5,
			},
			valid: false,
		},
		// this line is used by starport scaffolding # types/genesis/testcase
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
