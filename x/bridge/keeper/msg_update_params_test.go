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
		sdkmath.NewInt(10_000_000_000),
		"0x0Ac72d9E87B39DAAa81e4F3F29Ce8c45B2bE5fA9",
		100,
		time.Now(),
		[]string{},
		2*time.Hour,
		sdkmath.LegacyMustNewDecFromStr("0.3"),
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
