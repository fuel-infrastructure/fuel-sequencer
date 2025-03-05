package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

var _ sdk.Msg = &MsgSkippedEventTx{}

func NewMsgSkippedEventTx(authority string, reasonForSkip string, ethBlockNumber uint64, ethLogIndex uint64, ethTxIndex uint64, ethTxHash string) *MsgSkippedEventTx {
	return &MsgSkippedEventTx{
		Authority:      authority,
		ReasonForSkip:  reasonForSkip,
		EthBlockNumber: ethBlockNumber,
		EthLogIndex:    ethLogIndex,
		EthTxIndex:     ethTxIndex,
		EthTxHash:      ethTxHash,
	}
}

// ValidateBasic for this message should be a no-op so that we definitely AnteHandle this message.
// Since we generate the MsgSkippedEventTx ourselves, we expect the message to be valid anyway.
func (*MsgSkippedEventTx) ValidateBasic() error {
	return nil
}

// RawTxBytes converts the message to a valid tx that can be injected into a block and produces a tx result.
// The sequence, presumed to be unique, ensures that the generated tx is unique and thus has a unique tx hash.
func (m *MsgSkippedEventTx) RawTxBytes(sequence uint64) ([]byte, error) {

	msgSkippedEventTxAny, err := codectypes.NewAnyWithValue(m)
	if err != nil {
		return nil, err
	}

	return utils.ValidRawTxBytesFromAnyMsgs([]*codectypes.Any{msgSkippedEventTxAny}, sequence)
}
