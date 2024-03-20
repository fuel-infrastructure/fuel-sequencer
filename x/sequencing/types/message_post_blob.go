package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgPostBlob{}

func NewMsgPostBlob(
	from string,
	topic []byte,
	order math.Int,
	data []byte,
) *MsgPostBlob {
	return &MsgPostBlob{
		From:  from,
		Topic: topic,
		Order: order,
		Data:  data,
	}
}

func (msg *MsgPostBlob) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}

	return nil
}
