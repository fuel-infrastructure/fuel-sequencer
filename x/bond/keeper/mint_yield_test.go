package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func TestMintYield(t *testing.T) {
	keeper, ctx := keepertest.BondKeeper(t)

	// Test case: no yield parameters set
	require.NoError(t, keeper.MintYield(ctx))
	require.Equal(t, int64(0), keeper.GetYieldMintHeight(ctx))

	// Set up test parameters
	baseTime := time.Now()
	future := baseTime.Add(time.Hour)
	recipient := keeper.GetAuthority()
	yieldAmount := sdkmath.NewInt(1000000)

	// Test case: yield time not reached
	params := types.NewParams(recipient, &future, yieldAmount)
	require.NoError(t, keeper.SetParams(ctx, params))
	require.NoError(t, keeper.MintYield(ctx))
	require.Equal(t, int64(0), keeper.GetYieldMintHeight(ctx))

	// Test case: yield already minted
	require.NoError(t, keeper.MintYield(ctx))
	require.Equal(t, ctx.BlockHeight(), keeper.GetYieldMintHeight(ctx))

	// Test case: zero yield amount
	params = types.NewParams(recipient, &future, sdkmath.ZeroInt())
	require.NoError(t, keeper.SetParams(ctx, params))
	require.NoError(t, keeper.MintYield(ctx))
}
