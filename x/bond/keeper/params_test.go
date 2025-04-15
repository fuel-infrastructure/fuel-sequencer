package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func TestSetGetParams(t *testing.T) {
	k, ctx := keepertest.BondKeeper(t)

	// First check default values
	defaultParams := types.DefaultParams()
	require.EqualValues(t, sdkmath.LegacyZeroDec(), defaultParams.Inflation)
	require.EqualValues(t, "", defaultParams.Authority)

	// Set non-default values
	params := types.DefaultParams()
	params.Inflation = sdkmath.LegacyNewDecWithPrec(5, 1) // 0.5
	params.Authority = "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"

	// Verify non-default values are different from default
	require.NotEqual(t, defaultParams.Inflation, params.Inflation)
	require.NotEqual(t, defaultParams.Authority, params.Authority)

	require.NoError(t, k.SetParams(ctx, params))

	// Verify non-default values were set correctly
	retrievedParams := k.GetParams(ctx)
	require.EqualValues(t, sdkmath.LegacyNewDecWithPrec(5, 1), retrievedParams.Inflation)
	require.EqualValues(t, "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu", retrievedParams.Authority)
}
