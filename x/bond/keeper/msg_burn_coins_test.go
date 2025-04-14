package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/mock"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func setupMsgBurnCoins(t testing.TB) (types.MsgServer, sdk.Context, *mock.MockBankKeeper) {
	k, ctx := keepertest.BondKeeper(t)
	mockBankKeeper := k.GetBankKeeper().(*mock.MockBankKeeper)
	return keeper.NewMsgServerImpl(k), sdk.UnwrapSDKContext(ctx), mockBankKeeper
}

func TestMsgBurnCoins(t *testing.T) {
	ms, ctx, mockBankKeeper := setupMsgBurnCoins(t)

	// Test cases
	tests := []struct {
		name      string
		msg       *types.MsgBurnCoins
		expErr    bool
		expErrMsg string
	}{
		{
			name: "valid burn coins",
			msg: types.NewMsgBurnCoins(
				"cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
				sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(100))),
			),
			expErr: false,
		},
		{
			name: "invalid sender address",
			msg: types.NewMsgBurnCoins(
				"invalid",
				sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(100))),
			),
			expErr:    true,
			expErrMsg: "invalid bech32 string",
		},
		{
			name: "zero coins",
			msg: types.NewMsgBurnCoins(
				"cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
				sdk.NewCoins(),
			),
			expErr:    true,
			expErrMsg: "coins cannot be zero",
		},
		{
			name: "negative coins",
			msg: types.NewMsgBurnCoins(
				"cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
				sdk.Coins{sdk.Coin{Denom: "ufuel", Amount: sdkmath.NewInt(-100)}},
			),
			expErr:    true,
			expErrMsg: "invalid coins",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.expErr {
				// Setup mock expectations for successful case
				sender, err := sdk.AccAddressFromBech32(tc.msg.Sender)
				require.NoError(t, err)
				mockBankKeeper.On("SendCoinsFromAccountToModule", testifymock.Anything, sender, types.ModuleName, tc.msg.Coins).Return(nil)
				mockBankKeeper.On("BurnCoins", testifymock.Anything, types.ModuleName, tc.msg.Coins).Return(nil)
			}

			_, err := ms.BurnCoins(ctx, tc.msg)

			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
