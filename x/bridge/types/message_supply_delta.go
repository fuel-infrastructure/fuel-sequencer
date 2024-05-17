package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgSupplyDelta{}

func NewMsgSupplyDelta(authority string) *MsgSupplyDelta {
	return &MsgSupplyDelta{
		Authority: authority,
	}
}

func (m *MsgSupplyDelta) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address (%s)", err)
	}

	return nil
}

// FromSdkTx extracts MsgSupplyDelta from an SDK transaction which is expected to contain just MsgSupplyDelta.
func (m *MsgSupplyDelta) FromSdkTx(tx sdk.Tx) error {

	// MsgSupplyDelta will contain only one message.
	msgs := tx.GetMsgs()
	if len(msgs) != 1 {
		return fmt.Errorf("expected 1 msg in MsgSupplyDelta raw bytes, got %d", len(msgs))
	}

	// If the message is not a MsgSupplyDelta return an error.
	msg := msgs[0]
	if sdk.MsgTypeURL(msg) != sdk.MsgTypeURL(&MsgSupplyDelta{}) {
		return fmt.Errorf("expected msg type URL %s, got %s", sdk.MsgTypeURL(&MsgSupplyDelta{}), sdk.MsgTypeURL(msg))
	}

	// If the message cannot be parsed into MsgSupplyDelta, this is a problem.
	msgSupplyDelta, ok := msg.(*MsgSupplyDelta)
	if !ok {
		return errors.New("could not parse message into MsgSupplyDelta")
	}

	*m = *msgSupplyDelta
	return nil
}
