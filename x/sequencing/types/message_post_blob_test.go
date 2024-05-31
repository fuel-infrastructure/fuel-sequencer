package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgPostBlob_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgPostBlob
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgPostBlob{
				From: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgPostBlob{
				From: sample.AccAddress(),
			},
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
