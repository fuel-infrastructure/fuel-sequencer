package types_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func TestMsgWithdrawToEthereum_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  types.MsgWithdrawToEthereum
		err  error
	}{
		{
			name: "valid msg withdraw to ethereum - from is bech32",
			msg: types.MsgWithdrawToEthereum{
				From:   testtypes.TestFrom1Seq,
				To:     testtypes.TestTo1,
				Amount: sdk.NewCoin("fuel", math.NewInt(200)),
			},
		},
		{
			name: "valid msg withdraw to ethereum - from is Hex",
			msg: types.MsgWithdrawToEthereum{
				From:   testtypes.TestFrom1,
				To:     testtypes.TestTo1,
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
				From: testtypes.TestFrom1,
				To:   "0xZYXb5d4c32345ced77393b3530b1eed0f346429d",
			},
			err: types.ErrInvalidEthAddress,
		},
		{
			name: "invalid amount",
			msg: types.MsgWithdrawToEthereum{
				From: testtypes.TestFrom1,
				To:   testtypes.TestTo1,
			},
			err: sdkerrors.ErrInvalidCoins,
		},
		{
			name: "invalid zero amount",
			msg: types.MsgWithdrawToEthereum{
				From:   testtypes.TestFrom1,
				To:     testtypes.TestTo1,
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
