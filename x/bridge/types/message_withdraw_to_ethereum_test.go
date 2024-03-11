package types_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/sample"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestMsgWithdrawToEthereum_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  types.MsgWithdrawToEthereum
		err  error
	}{
		{
			name: "valid msg withdraw to ethereum",
			msg: types.MsgWithdrawToEthereum{
				From:   sample.AccAddress(),
				To:     "0x95222290dd7278aa3ddd389cc1e1d165cc4bafe5",
				Amount: sdk.NewCoin("fuel", math.NewInt(200)),
			},
		},
		{
			name: "invalid from address",
			msg: types.MsgWithdrawToEthereum{
				From: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid to address",
			msg: types.MsgWithdrawToEthereum{
				From: sample.AccAddress(),
				To:   "0xZYXb5d4c32345ced77393b3530b1eed0f346429d",
			},
			err: types.ErrInvalidEthAddress,
		},
		{
			name: "invalid amount",
			msg: types.MsgWithdrawToEthereum{
				From: sample.AccAddress(),
				To:   "0x95222290dd7278aa3ddd389cc1e1d165cc4bafe5",
			},
			err: sdkerrors.ErrInvalidCoins,
		},
		{
			name: "invalid zero amount",
			msg: types.MsgWithdrawToEthereum{
				From:   sample.AccAddress(),
				To:     "0x95222290dd7278aa3ddd389cc1e1d165cc4bafe5",
				Amount: sdk.NewCoin("fuel", math.ZeroInt()),
			},
			err: sdkerrors.ErrInvalidCoins,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
