package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

var _ sdk.Msg = &MsgWithdrawToEthereum{}

func NewMsgWithdrawToEthereum(from string, to string, amount sdk.Coin) *MsgWithdrawToEthereum {
	return &MsgWithdrawToEthereum{
		From:   from,
		To:     to,
		Amount: amount,
	}
}

func (msg *MsgWithdrawToEthereum) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid from address (%s)", err)
	}

	// TODO: We might want to verify checksum of address
	if !common.IsHexAddress(msg.To) {
		return ErrInvalidEthAddress.Wrapf("invalid Ethereum to address format (%s)", msg.To)
	}

	if !msg.Amount.IsValid() || msg.Amount.Amount.IsZero() {
		return sdkerrors.ErrInvalidCoins.Wrapf("amount must be a valid, non-zero value")
	}

	return nil
}
