package bridge_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	bridge "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/module"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		SupplyDeltaInfo: &types.SupplyDeltaInfo{
			LastSupply: math.NewInt(99),
			Offset:     math.NewInt(123),
			ToReport:   math.NewInt(87),
		},

		LastEthereumNonce:        math.NewInt(75),
		LastEthereumBlockSynced:  13,
		EthereumEventIndexOffset: 55,
		LastEthBlockUpdateTime:   time.Now().Round(0),
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.BridgeKeeper(t)
	bridge.InitGenesis(ctx, k, genesisState)
	got := bridge.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	require.Equal(t, genesisState.Params, got.Params)
	require.Equal(t, genesisState.SupplyDeltaInfo, got.SupplyDeltaInfo)
	require.Equal(t, genesisState.LastEthereumNonce, got.LastEthereumNonce)
	require.Equal(t, genesisState.LastEthereumBlockSynced, got.LastEthereumBlockSynced)
	require.Equal(t, genesisState.EthereumEventIndexOffset, got.EthereumEventIndexOffset)
	require.Equal(t, genesisState.LastEthBlockUpdateTime, got.LastEthBlockUpdateTime)
	require.Equal(t, genesisState.LastConsensusTxsSequence, got.LastConsensusTxsSequence)
	// this line is used by starport scaffolding # genesis/test/assert
}
