package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

func TestMsgUpdateParams(t *testing.T) {
	k, ms, ctx := setupMsgServer(t)
	defaultParams := types.DefaultParams()
	nonDefaultParams := types.NewParams(1000, 10)
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
			name: "all good with default params",
			input: &types.MsgUpdateParams{
				Authority: k.GetAuthority(),
				Params:    defaultParams,
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
