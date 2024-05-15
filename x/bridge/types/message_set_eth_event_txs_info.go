package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &EthEventsTx{}

func NewMsgSetEthEventTxsIndex(
	authority string, numInjectedEvents uint64, newEthereumBlock bool, blockNumber uint64,
) *EthEventsTx {
	return &EthEventsTx{
		Authority:         authority,
		NumInjectedEvents: numInjectedEvents,
		NewEthereumBlock:  newEthereumBlock,
		BlockNumber:       blockNumber,
	}
}

func (msg *EthEventsTx) ValidateBasic() error {

	// Note: sufficient validation already done during injection and in the message handler.

	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}
	return nil
}
