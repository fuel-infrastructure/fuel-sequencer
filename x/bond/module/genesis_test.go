package bond_test

import (
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	bond "github.com/fuel-infrastructure/fuel-sequencer/x/bond/module"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func TestGenesis(t *testing.T) {
	// Create test dependencies
	k, ctx, cdc, mockAccountKeeper, mockBankKeeper := keeper.BondKeeperWithDependencies(t)

	// Create test module
	appModule := bond.NewAppModule(
		cdc,
		k,
		mockAccountKeeper,
		mockBankKeeper,
	)

	// Test module name
	require.Equal(t, "bond", appModule.Name())

	// Test module version
	require.Equal(t, uint64(1), appModule.ConsensusVersion())

	// Set up mock expectations
	mockModuleAccount := authtypes.NewEmptyModuleAccount(types.ModuleName)
	mockAccountKeeper.EXPECT().
		GetModuleAccount(gomock.Any(), types.ModuleName).
		Return(mockModuleAccount).
		Times(1)

	// Test module genesis
	require.NotPanics(t, func() {
		appModule.InitGenesis(ctx, cdc, cdc.MustMarshalJSON(types.DefaultGenesis()))
	})

	// Test module export
	exported := appModule.ExportGenesis(ctx, cdc)
	var exportedState types.GenesisState
	cdc.MustUnmarshalJSON(exported, &exportedState)
	require.Equal(t, types.DefaultGenesis().Params, exportedState.Params)
}
