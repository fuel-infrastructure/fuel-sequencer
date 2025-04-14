package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func TestNewMsgBurnCoins(t *testing.T) {
	testCases := []struct {
		name   string
		sender string
		coins  sdk.Coins
		expErr bool
	}{
		{
			name:   "valid message",
			sender: "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
			coins:  sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(100))),
			expErr: false,
		},
		{
			name:   "invalid sender",
			sender: "invalid",
			coins:  sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(100))),
			expErr: true,
		},
		{
			name:   "zero coins",
			sender: "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
			coins:  sdk.NewCoins(),
			expErr: true,
		},
		{
			name:   "negative coins",
			sender: "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
			coins:  sdk.Coins{sdk.Coin{Denom: "ufuel", Amount: sdkmath.NewInt(-100)}},
			expErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgBurnCoins(tc.sender, tc.coins)
			err := msg.ValidateBasic()
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.sender, msg.Sender)
				require.Equal(t, tc.coins, msg.Coins)
			}
		})
	}
}

func TestMsgBurnCoins_GetSigners(t *testing.T) {
	testCases := []struct {
		name   string
		sender string
		expErr bool
	}{
		{
			name:   "valid sender",
			sender: "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
			expErr: false,
		},
		{
			name:   "invalid sender",
			sender: "invalid",
			expErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgBurnCoins(tc.sender, sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(100))))
			if tc.expErr {
				require.Panics(t, func() { msg.GetSigners() })
			} else {
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, tc.sender, signers[0].String())
			}
		})
	}
}
