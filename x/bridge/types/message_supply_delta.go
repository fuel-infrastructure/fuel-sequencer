package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgSupplyDelta{}

func NewMsgSupplyDelta(authority string) *MsgSupplyDelta {
	return &MsgSupplyDelta{
		Authority: authority,
	}
}

func (msg *MsgSupplyDelta) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	// TODO: validate other fields if necessary

	return nil
}
