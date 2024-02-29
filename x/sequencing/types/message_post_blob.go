package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgPostBlob{}

func NewMsgPostBlob(from string, topic string, order string, data []byte) *MsgPostBlob {
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
