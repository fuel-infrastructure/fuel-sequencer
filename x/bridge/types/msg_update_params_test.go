package types_test

import (
	sdkmath "cosmossdk.io/math"
	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestMsgUpdateParams_ValidateBasic(t *testing.T) {
	k, _ := keepertest.BridgeKeeper(t)
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

	testCases := []struct {
		name            string
		msgUpdateParams *types.MsgUpdateParams
		expErr          bool
		expErrMsg       string
	}{
		{
			name: "good with non-default params",
			msgUpdateParams: &types.MsgUpdateParams{
				Authority: k.GetAuthority(),
				Params:    nonDefaultParams,
			},
			expErr: false,
		},
		{
			name: "not good with default params due to vestingStartTime",
			msgUpdateParams: &types.MsgUpdateParams{
				Authority: k.GetAuthority(),
				Params:    defaultParams,
			},
			expErr:    true,
			expErrMsg: "vesting start time must be set and cannot be the zero value",
		},
		{
			name: "not good with due to wrong authority",
			msgUpdateParams: &types.MsgUpdateParams{
				Authority: "incorrect",
				Params:    nonDefaultParams,
			},
			expErr:    true,
			expErrMsg: "invalid authority address: decoding bech32 failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Validate the message first
			err := tc.msgUpdateParams.ValidateBasic()
			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
