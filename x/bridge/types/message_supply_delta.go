package types

import (
	"errors"
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

var _ sdk.Msg = &MsgSupplyDelta{}

func NewMsgSupplyDelta(authority string) *MsgSupplyDelta {
	return &MsgSupplyDelta{
		Authority: authority,
	}
}

// ValidateBasic for this message should be a no-op so that we definitely AnteHandle this message.
// Since we generate the MsgSupplyDelta ourselves, we expect the message to be valid anyway.
func (*MsgSupplyDelta) ValidateBasic() error {
	return nil
}

// RawTxBytes converts the message to a valid tx that can be injected into a block and produces a tx result.
// The sequence, presumed to be unique, ensures that the generated tx is unique and thus has a unique tx hash.
func (m *MsgSupplyDelta) RawTxBytes(sequence uint64) ([]byte, error) {

	// Construct Any from message.
	msgSupplyDeltaAny, err := codectypes.NewAnyWithValue(m)
	if err != nil {
		return nil, err
	}

	return utils.ValidRawTxBytesFromAnyMsgs([]*codectypes.Any{msgSupplyDeltaAny}, sequence)
}

// FromSdkTx extracts MsgSupplyDelta from an SDK transaction which is expected to contain just MsgSupplyDelta.
func (m *MsgSupplyDelta) FromSdkTx(tx sdk.Tx) error {

	if m == nil {
		return fmt.Errorf("expected non-nil MsgSupplyDelta receiver")
	}

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
