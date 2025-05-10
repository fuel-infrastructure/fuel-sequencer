package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

const (
	testAddr = "fuelsequencer1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5gjkhx7"
)

func TestSetGetParams(t *testing.T) {
	k, ctx := keepertest.BondKeeper(t)

	// First check default values
	defaultParams := types.DefaultParams()
	require.EqualValues(t, "", defaultParams.YieldRecipient)
	require.EqualValues(t, (*time.Time)(nil), defaultParams.YieldTime)
	require.EqualValues(t, sdkmath.ZeroInt(), defaultParams.YieldAmount)

	// Set non-default values
	baseTime := time.Now()
	future := baseTime.Add(time.Hour)
	params := types.DefaultParams()
	params.YieldRecipient = testAddr
	params.YieldTime = &future
	params.YieldAmount = sdkmath.NewInt(1000000)

	// Verify non-default values are different from default
	require.NotEqual(t, defaultParams.YieldRecipient, params.YieldRecipient)
	require.NotEqual(t, defaultParams.YieldTime, params.YieldTime)
	require.NotEqual(t, defaultParams.YieldAmount, params.YieldAmount)

	require.NoError(t, k.SetParams(ctx, params))

	// Verify non-default values were set correctly
	retrievedParams := k.GetParams(ctx)
	require.EqualValues(t, testAddr, retrievedParams.YieldRecipient)
	require.EqualValues(t, future.Unix(), retrievedParams.YieldTime.Unix())
	require.EqualValues(t, sdkmath.NewInt(1000000), retrievedParams.YieldAmount)
}

func (suite *KeeperTestSuite) TestParams() {
	params := types.DefaultParams()
	params.YieldRecipient = testAddr

	// Set params
	err := suite.App.BondKeeper.SetParams(suite.Ctx(), params)
	require.NoError(suite.T(), err)

	// Get params
	retrievedParams := suite.App.BondKeeper.GetParams(suite.Ctx())
	require.EqualValues(suite.T(), testAddr, retrievedParams.YieldRecipient)
}
