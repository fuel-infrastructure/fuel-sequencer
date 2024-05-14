package types

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

var _ sdk.Msg = &MsgDepositFromEthereum{}

func NewMsgDepositFromEthereum(depositor string, recipient string, amount string, lockup string) *MsgDepositFromEthereum {
	return &MsgDepositFromEthereum{
		Depositor: depositor,
		Recipient: recipient,
		Amount:    amount,
		Lockup:    lockup,
	}
}

// ValidateBasic for MsgDepositFromEthereum is essentially a copy of DepositEvent.ValidateBasic.
func (msg *MsgDepositFromEthereum) ValidateBasic() error {
	// TODO: Consider clearing out ValidateBasic since this might cause the deposit event to get skipped!

	// Error if the receiver is nil
	if msg == nil {
		return errors.New("MsgDepositFromEthereum is nil")
	}

	// Check that Depositor is a valid hex address
	if !common.IsHexAddress(msg.Depositor) {
		return errors.New("depositor is not a valid hex address")
	}

	// Check that Recipient is either a valid Sequencer or hex address.
	_, err := sdk.AccAddressFromBech32(msg.Recipient)
	if err != nil && !common.IsHexAddress(msg.Recipient) {
		return fmt.Errorf("recipient is not a valid Bech32 or Hex address")
	}

	// Check that the Lockup can be converted from a string to sdk.Int
	if _, success := sdkmath.NewIntFromString(msg.Lockup); !success {
		return errors.New("could not convert lockup to a valid sdk.Int")
	}

	// Check that the Amount can be converted from a string to sdk.Int and is bigger than zero
	amount, success := sdkmath.NewIntFromString(msg.Amount)
	if !success {
		return errors.New("could not convert amount to a valid sdk.Int")
	}
	if amount.IsZero() {
		return errors.New("amount must be bigger than zero")
	}

	return nil
}
