package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func TestMsgUpdateParams(t *testing.T) {
	k, ms, ctx := setupMsgServer(t)
	defaultParams := types.DefaultParams()
	nonDefaultParams := types.NewParams(
		"ufuel",
		"0x0Ac72d9E87B39DAAa81e4F3F29Ce8c45B2bE5fA9",
		[]string{"/cosmos.bank.v1beta1.MsgSend"},
		100,
		time.Now(),
		[]string{},
		2*time.Hour,
		6144,
		sdkmath.LegacyMustNewDecFromStr("0.3"),
		2,
	)
	require.NoError(t, k.SetParams(ctx, defaultParams))
	wctx := sdk.UnwrapSDKContext(ctx)

	// default params
	testCases := []struct {
		name      string
		input     *types.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			input: &types.MsgUpdateParams{
				Authority: "invalid",
				Params:    defaultParams,
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "send enabled param",
			input: &types.MsgUpdateParams{
				Authority: k.GetAuthority(),
				Params:    types.Params{},
			},
			expErr: false,
		},
		{
			name: "not good with default params due to vestingStartTime",
			input: &types.MsgUpdateParams{
				Authority: k.GetAuthority(),
				Params:    defaultParams,
			},
			expErr:    true,
			expErrMsg: "vesting start time must be set and cannot be the zero value",
		},
		{
			name: "all good with non default params",
			input: &types.MsgUpdateParams{
				Authority: k.GetAuthority(),
				Params:    nonDefaultParams,
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// Validate the message first
			err := tc.input.ValidateBasic()
			if err != nil && tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				_, err = ms.UpdateParams(wctx, tc.input)

				if tc.expErr {
					require.Error(t, err)
					require.Contains(t, err.Error(), tc.expErrMsg)
				} else {
					require.NoError(t, err)
				}
			}
		})
	}
}
