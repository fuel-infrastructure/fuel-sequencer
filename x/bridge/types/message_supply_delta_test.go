package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgSupplyDelta_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgSupplyDelta
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgSupplyDelta{
				Authority: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgSupplyDelta{
				Authority: sample.AccAddress(),
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
