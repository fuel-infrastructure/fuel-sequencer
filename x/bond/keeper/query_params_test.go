package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := keepertest.BondKeeper(t)

	testCases := []struct {
		name      string
		setup     func() types.Params
		expParams types.Params
		expErr    bool
	}{
		{
			name: "default params",
			setup: func() types.Params {
				params := types.DefaultParams()
				require.NoError(t, keeper.SetParams(ctx, params))
				return params
			},
			expParams: types.DefaultParams(),
			expErr:    false,
		},
		{
			name: "custom params",
			setup: func() types.Params {
				fixedTime := time.Now().Add(time.Hour)
				params := types.NewParams(
					keeper.GetAuthority(),
					&fixedTime,
					sdkmath.NewInt(1000000),
				)
				require.NoError(t, keeper.SetParams(ctx, params))
				return params
			},
			expParams: func() types.Params {
				fixedTime := time.Now().Add(time.Hour)
				return types.NewParams(
					keeper.GetAuthority(),
					&fixedTime,
					sdkmath.NewInt(1000000),
				)
			}(),
			expErr: false,
		},
		{
			name: "empty params",
			setup: func() types.Params {
				params := types.Params{}
				require.NoError(t, keeper.SetParams(ctx, params))
				return params
			},
			expParams: types.Params{},
			expErr:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			tc.setup()

			// Execute query
			response, err := keeper.Params(ctx, &types.QueryParamsRequest{})

			// Verify
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				// Compare params without considering time zone
				if tc.expParams.YieldTime != nil {
					require.Equal(t, tc.expParams.YieldTime.UTC().Unix(), response.Params.YieldTime.UTC().Unix())
				} else {
					require.Nil(t, response.Params.YieldTime)
				}
				require.Equal(t, tc.expParams.YieldRecipient, response.Params.YieldRecipient)
				// Compare big.Int values
				if tc.expParams.YieldAmount.IsNil() && response.Params.YieldAmount.IsZero() {
					// treat nil and zero as equivalent
					return
				}
				if tc.expParams.YieldAmount.IsNil() {
					require.True(t, response.Params.YieldAmount.IsNil())
				} else {
					require.Equal(t, tc.expParams.YieldAmount, response.Params.YieldAmount)
				}
			}
		})
	}
}

func TestParamsQuery_InvalidRequest(t *testing.T) {
	keeper, ctx := keepertest.BondKeeper(t)

	// Test with nil request
	response, err := keeper.Params(ctx, nil)
	require.Error(t, err)
	require.Nil(t, response)
}
