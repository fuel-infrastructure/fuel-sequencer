package types_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	testutiltypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func TestMsgUpdateParams_ValidateBasic(t *testing.T) {
	k, _ := keepertest.BridgeKeeper(t)

	defaultParams := types.DefaultParams()

	// Default params but with values that will be considered valid
	validDefaultParams := defaultParams
	validDefaultParams.VestingStartTime = time.Now()
	validDefaultParams.BridgeDenomTotalSupply = testutiltypes.TestBridgeDenomTotalSupply

	// Params with default (invalid) BridgeDenomTotalSupply
	paramsDefaultBridgeDenomTotalSupply := validDefaultParams
	paramsDefaultBridgeDenomTotalSupply.BridgeDenomTotalSupply = defaultParams.BridgeDenomTotalSupply

	// Params with default (invalid) VestingStartTime
	paramsDefaultVestingStartTime := validDefaultParams
	paramsDefaultVestingStartTime.VestingStartTime = defaultParams.VestingStartTime

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
			name: "not good with default VestingStartTime",
			msgUpdateParams: &types.MsgUpdateParams{
				Authority: k.GetAuthority(),
				Params:    paramsDefaultVestingStartTime,
			},
			expErr:    true,
			expErrMsg: "vesting start time must be set and cannot be the zero value",
		},
		{
			name: "not good with default BridgeDenomTotalSupply",
			msgUpdateParams: &types.MsgUpdateParams{
				Authority: k.GetAuthority(),
				Params:    paramsDefaultBridgeDenomTotalSupply,
			},
			expErr:    true,
			expErrMsg: "bridge denom total supply must be positive, got: 0",
		},
		{
			name: "not good due to wrong authority",
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
