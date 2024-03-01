package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgWithdrawToEthereum{}

func NewMsgWithdrawToEthereum(from string, nonce string, to string, amount sdk.Coin) *MsgWithdrawToEthereum {
	return &MsgWithdrawToEthereum{
		From:   from,
		To:     to,
		Amount: amount,
	}
}

func (msg *MsgWithdrawToEthereum) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}

	// TODO: validate other fields if necessary

	return nil
}
