package types

import (
	"errors"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	// NOTE: we intentionally skip the validation of the Depositor and Recipient since in the message handler we still
	// want to mint the deposited tokens even if we do not know who they are coming from or who they are going to.

	// Error if the receiver is nil
	if msg == nil {
		return errors.New("MsgDepositFromEthereum is nil")
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
