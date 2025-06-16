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

func TestParams(t *testing.T) {
	k, ctx := keepertest.BondKeeper(t)
	baseTime := time.Now()
	future := baseTime.Add(time.Hour)

	testCases := []struct {
		name        string
		params      types.Params
		expectError bool
		check       func(t *testing.T, params types.Params)
	}{
		{
			name:   "default params",
			params: types.DefaultParams(),
			check: func(t *testing.T, params types.Params) {
				require.EqualValues(t, "", params.YieldRecipient)
				require.EqualValues(t, (*time.Time)(nil), params.YieldTime)
				require.EqualValues(t, sdkmath.ZeroInt(), params.YieldAmount)
			},
		},
		{
			name: "set and get valid params",
			params: types.Params{
				YieldRecipient: testAddr,
				YieldTime:      &future,
				YieldAmount:    sdkmath.NewInt(1000000),
			},
			check: func(t *testing.T, params types.Params) {
				require.EqualValues(t, testAddr, params.YieldRecipient)
				require.EqualValues(t, future.Unix(), params.YieldTime.Unix())
				require.EqualValues(t, sdkmath.NewInt(1000000), params.YieldAmount)
			},
		},
		{
			name: "invalid recipient address",
			params: types.Params{
				YieldRecipient: "invalid_address",
				YieldTime:      &future,
				YieldAmount:    sdkmath.NewInt(1000000),
			},
			expectError: true,
		},
		{
			name: "negative yield amount",
			params: types.Params{
				YieldRecipient: testAddr,
				YieldTime:      &future,
				YieldAmount:    sdkmath.NewInt(-1000),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := k.SetParams(ctx, tc.params)
			if tc.expectError {
				require.Error(t, err)
				if tc.name == "invalid recipient address" {
					require.Contains(t, err.Error(), "invalid yield recipient address")
				}
				if tc.name == "negative yield amount" {
					require.Contains(t, err.Error(), "yield amount cannot be negative")
				}
			} else {
				require.NoError(t, err)
				retrievedParams := k.GetParams(ctx)
				if tc.check != nil {
					tc.check(t, retrievedParams)
				}
			}
		})
	}
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
