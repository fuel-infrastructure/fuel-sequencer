package keeper_test

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func TestMsgUpdateParams(t *testing.T) {
	k, ms, ctx := setupMsgServer(t)
	params := types.DefaultParams()
	require.NoError(t, k.SetParams(ctx, params))
	wctx := sdk.UnwrapSDKContext(ctx)

	// Get the expected authority from the keeper
	expectedAuthority := k.GetAuthority()

	// Helper function to create error message for authority validation
	authorityErrMsg := func(got string) string {
		return fmt.Sprintf("invalid authority; expected %s, got %s: invalid signer", expectedAuthority, got)
	}

	future := time.Now().Add(time.Hour)

	testCases := []struct {
		name      string
		input     *types.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		// Authority validation test cases
		{
			name: "should fail when authority is malformed",
			input: &types.MsgUpdateParams{
				Authority: "invalid",
				Params:    params,
			},
			expErr:    true,
			expErrMsg: authorityErrMsg("invalid"),
		},
		{
			name: "should fail when authority is empty",
			input: &types.MsgUpdateParams{
				Authority: "",
				Params:    params,
			},
			expErr:    true,
			expErrMsg: authorityErrMsg(""),
		},
		{
			name: "should fail when authority is different from expected",
			input: &types.MsgUpdateParams{
				Authority: "fuelsequencer1wrongaddress",
				Params:    params,
			},
			expErr:    true,
			expErrMsg: authorityErrMsg("fuelsequencer1wrongaddress"),
		},

		// Valid parameter test cases
		{
			name: "should succeed with empty params",
			input: &types.MsgUpdateParams{
				Authority: expectedAuthority,
				Params:    types.Params{},
			},
			expErr: false,
		},
		{
			name: "should succeed with default params",
			input: &types.MsgUpdateParams{
				Authority: expectedAuthority,
				Params:    params,
			},
			expErr: false,
		},
		{
			name: "should succeed with custom yield params",
			input: &types.MsgUpdateParams{
				Authority: expectedAuthority,
				Params: types.NewParams(
					expectedAuthority,
					&future,
					sdkmath.NewInt(1000000),
				),
			},
			expErr: false,
		},
		{
			name: "should fail with invalid recipient address",
			input: &types.MsgUpdateParams{
				Authority: expectedAuthority,
				Params: types.NewParams(
					"invalid",
					&future,
					sdkmath.NewInt(1000000),
				),
			},
			expErr:    true,
			expErrMsg: "invalid yield recipient address: decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name: "should fail with negative yield amount",
			input: &types.MsgUpdateParams{
				Authority: expectedAuthority,
				Params: types.NewParams(
					expectedAuthority,
					&future,
					sdkmath.NewInt(-1),
				),
			},
			expErr:    true,
			expErrMsg: "yield amount cannot be negative: -1",
		},
		{
			name: "should fail with past yield time",
			input: &types.MsgUpdateParams{
				Authority: expectedAuthority,
				Params: types.NewParams(
					expectedAuthority,
					&time.Time{}, // zero time is in the past
					sdkmath.NewInt(1000000),
				),
			},
			expErr:    true,
			expErrMsg: "yield time cannot be in the past",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ms.UpdateParams(wctx, tc.input)

			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
				// Verify params were updated
				updatedParams := k.GetParams(ctx)
				// Compare params without considering time zone
				if tc.input.Params.YieldTime != nil {
					require.True(t, tc.input.Params.YieldTime.UTC().Equal(updatedParams.YieldTime.UTC()))
				} else {
					require.Nil(t, updatedParams.YieldTime)
				}
				require.Equal(t, tc.input.Params.YieldRecipient, updatedParams.YieldRecipient)
				// Compare big.Int values
				if tc.input.Params.YieldAmount.IsNil() && updatedParams.YieldAmount.IsZero() {
					// treat nil and zero as equivalent
					return
				}
				if tc.input.Params.YieldAmount.IsNil() {
					require.True(t, updatedParams.YieldAmount.IsNil())
				} else {
					require.Equal(t, tc.input.Params.YieldAmount, updatedParams.YieldAmount)
				}
			}
		})
	}
}
