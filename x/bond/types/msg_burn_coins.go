package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgBurnCoins{}

// NewMsgBurnCoins creates a new MsgBurnCoins instance
func NewMsgBurnCoins(sender string, coins sdk.Coins) *MsgBurnCoins {
	return &MsgBurnCoins{
		Sender: sender,
		Coins:  coins,
	}
}

// ValidateBasic performs basic validation of the message
func (msg *MsgBurnCoins) ValidateBasic() error {
	if !msg.Coins.IsValid() {
		return fmt.Errorf("invalid coins: %s", msg.Coins)
	}

	if msg.Coins.IsZero() {
		return fmt.Errorf("coins cannot be zero")
	}

	return nil
}

// GetSigners returns the required signers of the message
func (msg *MsgBurnCoins) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}
