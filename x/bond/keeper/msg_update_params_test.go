package keeper_test

import (
	"fmt"
	"testing"

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
			name: "should succeed with custom inflation rate",
			input: &types.MsgUpdateParams{
				Authority: expectedAuthority,
				Params: types.NewParams(
					sdkmath.LegacyNewDecWithPrec(5, 1), // 0.5
					expectedAuthority,
				),
			},
			expErr: false,
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
			}
		})
	}
}
