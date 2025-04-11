package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMsgBurnCoins_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgBurnCoins
		wantErr bool
	}{
		{
			name: "valid message",
			msg: NewMsgBurnCoins(
				"cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
				sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(100))),
			),
			wantErr: false,
		},
		{
			name: "invalid sender address",
			msg: NewMsgBurnCoins(
				"invalid",
				sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(100))),
			),
			wantErr: true,
		},
		{
			name: "invalid coins - zero amount",
			msg: NewMsgBurnCoins(
				"cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
				sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(0))),
			),
			wantErr: true,
		},
		{
			name: "invalid coins - negative amount",
			msg: NewMsgBurnCoins(
				"cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
				sdk.Coins{sdk.Coin{Denom: "ufuel", Amount: sdkmath.NewInt(-100)}},
			),
			wantErr: true,
		},
		{
			name: "invalid coins - empty coins",
			msg: NewMsgBurnCoins(
				"cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
				sdk.NewCoins(),
			),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgBurnCoins_GetSigners(t *testing.T) {
	sender := "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"
	msg := NewMsgBurnCoins(
		sender,
		sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(100))),
	)

	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, sender, signers[0].String())
}
