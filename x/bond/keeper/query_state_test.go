package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func TestStateQuery_DefaultState(t *testing.T) {
	keeper, ctx := keepertest.BondKeeper(t)

	// Setup
	state := types.DefaultState()
	require.NoError(t, keeper.SetState(ctx, state))

	// Execute query
	response, err := keeper.State(ctx, &types.QueryStateRequest{})

	// Verify
	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.State)
	require.Equal(t, int64(0), response.State.YieldMintHeight, "default yield mint height should be 0")

	// Verify state is queryable multiple times
	response2, err := keeper.State(ctx, &types.QueryStateRequest{})
	require.NoError(t, err)
	require.NotNil(t, response2)
	require.Equal(t, response.State, response2.State, "state should be consistent across queries")
}

func TestStateQuery_CustomState(t *testing.T) {
	keeper, ctx := keepertest.BondKeeper(t)

	// Setup
	state := types.NewState(100)
	require.NoError(t, keeper.SetState(ctx, state))

	// Execute query
	response, err := keeper.State(ctx, &types.QueryStateRequest{})

	// Verify
	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.State)
	require.Equal(t, int64(100), response.State.YieldMintHeight, "yield mint height should match custom value")
}

func TestStateQuery_InvalidRequest(t *testing.T) {
	keeper, ctx := keepertest.BondKeeper(t)

	// Test with nil request
	response, err := keeper.State(ctx, nil)
	require.Error(t, err)
	require.Nil(t, response)
}
