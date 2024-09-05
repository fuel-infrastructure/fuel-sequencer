package types

import (
	"errors"
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

var _ sdk.Msg = &MsgIndex{}

func NewMsgIndex(authority string, numInjectedEventTxs uint64, newEthereumBlock bool, blockNumber uint64) *MsgIndex {
	return &MsgIndex{
		Authority:           authority,
		NumInjectedEventTxs: numInjectedEventTxs,
		NewEthereumBlock:    newEthereumBlock,
		BlockNumber:         blockNumber,
	}
}

// ValidateBasic for this message should be a no-op so that we definitely AnteHandle this message.
// Since we generate the MsgIndex ourselves, we expect the message to be valid anyway.
func (*MsgIndex) ValidateBasic() error {
	return nil
}

// ValidateBeforeProcessing performs some state-based checks on MsgIndex before it is officially processed.
func (m *MsgIndex) ValidateBeforeProcessing(lastBlockSynced, eventIndexOffset uint64) error {

	// If we have an offset, we expect at least one new event.
	//
	// The cases where we receive NO events are:
	//
	// - New Ethereum block with no events - but since the offset is non-zero, we know that the current Ethereum block
	//   being synced still has more events for us to consume, so this case is invalid.
	// - No new Ethereum block - but since the offset is non-zero, we know that the current Ethereum block exists, and
	//   we expect to receive the remaining events. The current block is a new Ethereum block, so this case is invalid.
	if eventIndexOffset > 0 && m.NumInjectedEventTxs == 0 {
		return fmt.Errorf(
			"expected at least 1 new event if offset is non-zero (%d), got MsgIndex (%s)",
			eventIndexOffset, m,
		)
	}

	// BlockNumber must be LastEthereumBlockSynced+1 since otherwise we're getting data for an Ethereum block
	// that we've already fully processed, or we're getting data for an Ethereum block that is in the future.
	expectedBlockNumber := lastBlockSynced + 1
	if m.BlockNumber != expectedBlockNumber {
		return fmt.Errorf(
			"expected block number %d, got %d in MsgIndex (%s)",
			expectedBlockNumber, m.BlockNumber, m,
		)
	}

	return nil
}

// NumberOfEventsWithMaxBytes calculates the number of events that can fit into the specified maxBytes. This considers
// the size of the MsgIndex as raw tx bytes and iterates over as many events as can fit into the specified maxBytes.
func (m *MsgIndex) NumberOfEventsWithMaxBytes(eventTxs [][]byte, maxBytes, indexSequence uint64) (int, error) {

	msgIndexRawTxBytes, err := m.RawTxBytes(indexSequence)
	if err != nil {
		return 0, err
	}

	// If the MsgIndex on its own cannot fit into the block, this is a major issue.
	msgIndexTxSize := utils.TxSize(msgIndexRawTxBytes)
	if msgIndexTxSize > maxBytes {
		return 0, fmt.Errorf(
			"could not fit MsgIndex of size %d in max bytes allocated for events %d", msgIndexTxSize, maxBytes,
		)
	}

	n := msgIndexTxSize
	for i, eventTx := range eventTxs {
		toAdd := utils.TxSize(eventTx)
		if n+toAdd > maxBytes {
			return i, nil
		}
		n += toAdd
	}

	return len(eventTxs), nil
}

// TrimEventsFromHead removes the first N events from the front of the list of events.
//
// An important check that it does is to ensure that if there are events, these cannot all be trimmed, otherwise the
// blockchain might get stuck injecting empty MsgIndex forever. At least one event must be kept if there are events.
func (m *MsgIndex) TrimEventsFromHead(eventTxs [][]byte, numEventsToTrim uint64) ([][]byte, error) {

	if numEventsToTrim == 0 {

		return eventTxs, nil // trim nothing

	} else if m.NumInjectedEventTxs > 0 && m.NumInjectedEventTxs == numEventsToTrim {

		// If we trim all the events from the transaction this is a problem because if we retry
		// at the next block, we expect the same to happen, and we will never inject the events.
		return nil, fmt.Errorf("cannot trim all %d events", m.NumInjectedEventTxs)

	} else if numEventsToTrim > m.NumInjectedEventTxs {

		// If we try to trim more events than there are, something is wrong.
		return nil, fmt.Errorf(
			"insufficient no of events, expected at least %d got %d",
			numEventsToTrim, m.NumInjectedEventTxs,
		)

	}

	// Note: There is no need to use utils.TxSize to calculate the size of MsgIndex because we are only checking the
	// size of the message and not how big the transaction containing the message would be.
	msgIndexSizeBefore := m.Size()
	eventTxs = eventTxs[numEventsToTrim:]
	m.NumInjectedEventTxs = uint64(len(eventTxs))

	// Check whether the modification of MsgIndex has increased its size to avoid unexpected behaviour
	if m.Size() > msgIndexSizeBefore {
		return nil, fmt.Errorf("unexpected increase of MsgIndex size from %d to %d", msgIndexSizeBefore, m.Size())
	}

	return eventTxs, nil
}

// KeepEventsFromHead keeps the first N events from the front of the list of events and trims the rest.
//
// An important check that it does is to ensure that if there are events, these cannot all be trimmed, otherwise the
// blockchain might get stuck injecting empty MsgIndex forever. At least one event must be kept if there are events.
func (m *MsgIndex) KeepEventsFromHead(
	eventTxs [][]byte, numEventsToKeep uint64,
) (newEventTxs [][]byte, trimmed uint64, err error) {

	if m.NumInjectedEventTxs == numEventsToKeep {

		return eventTxs, 0, nil // keep all

	} else if m.NumInjectedEventTxs > 0 && numEventsToKeep == 0 {

		// If we trim all the events from the transaction this is a problem because if we retry
		// at the next block, we expect the same to happen, and we will never inject the events.
		return nil, 0, fmt.Errorf("cannot trim all %d events", m.NumInjectedEventTxs)

	} else if m.NumInjectedEventTxs < numEventsToKeep {

		// If we try to trim more events than there are, something is wrong.
		return nil, 0, fmt.Errorf(
			"insufficient no of events, expected at least %d got %d",
			numEventsToKeep, m.NumInjectedEventTxs,
		)

	}

	// Note: There is no need to use utils.TxSize to calculate the size of MsgIndex because we are only checking the
	// size of the message and not how big the transaction containing the message would be.
	msgIndexSizeBefore := m.Size()
	eventTxs = eventTxs[:numEventsToKeep]
	trimmed = m.NumInjectedEventTxs - numEventsToKeep
	m.NumInjectedEventTxs = uint64(len(eventTxs))
	m.NewEthereumBlock = false

	// Check whether the modification of MsgIndex has increased its size to avoid unexpected behaviour
	if m.Size() > msgIndexSizeBefore {
		return nil, 0, fmt.Errorf("unexpected increase of MsgIndex size from %d to %d", msgIndexSizeBefore, m.Size())
	}

	return eventTxs, trimmed, nil
}

// RawTxBytes converts the message to a valid tx that can be injected into a block and produces a tx result.
func (m *MsgIndex) RawTxBytes(sequence uint64) ([]byte, error) {

	msgIndexAny, err := codectypes.NewAnyWithValue(m)
	if err != nil {
		return nil, err
	}

	msgIndexBz, err := utils.ValidRawTxBytesFromAnyMsgs([]*codectypes.Any{msgIndexAny}, sequence)
	if err != nil {
		return nil, err
	}

	return msgIndexBz, nil
}

// FromSdkTx extracts MsgIndex from an SDK transaction which is expected to contain just MsgIndex.
func (m *MsgIndex) FromSdkTx(tx sdk.Tx) error {

	if m == nil {
		return fmt.Errorf("expected non-nil MsgIndex receiver")
	}

	// MsgIndex will contain only one message.
	msgs := tx.GetMsgs()
	if len(msgs) != 1 {
		return fmt.Errorf("expected 1 msg in MsgIndex raw bytes, got %d", len(msgs))
	}

	// If the message is not a MsgIndex return an error.
	msg := msgs[0]
	if sdk.MsgTypeURL(msg) != sdk.MsgTypeURL(&MsgIndex{}) {
		return fmt.Errorf("expected msg type URL %s, got %s", sdk.MsgTypeURL(&MsgIndex{}), sdk.MsgTypeURL(msg))
	}

	// If the message cannot be parsed into MsgIndex, this is a problem.
	msgIndex, ok := msg.(*MsgIndex)
	if !ok {
		return errors.New("could not parse message into MsgIndex")
	}

	*m = *msgIndex
	return nil
}

// FromRawTxBytes extracts MsgIndex from raw transaction bytes.
func (m *MsgIndex) FromRawTxBytes(bz []byte, decoder sdk.TxDecoder) error {

	if m == nil {
		return fmt.Errorf("expected non-nil MsgIndex receiver")
	}

	tx, err := decoder(bz)
	if err != nil {
		return err
	}

	return m.FromSdkTx(tx)
}

// IsFullEthereumSyncing returns true if an Ethereum block was fully consumed by the Sequencer.
func (m *MsgIndex) IsFullEthereumSyncing() bool {
	return m.NewEthereumBlock
}

// IsPartialEthereumSyncing returns true if an Ethereum block was not fully consumed by the Sequencer
// and that the events in the Ethereum block will be split into multiple Sequencer blocks.
func (m *MsgIndex) IsPartialEthereumSyncing() bool {
	return !m.NewEthereumBlock && m.NumInjectedEventTxs > 0
}

// NoEthereumSyncing returns true if there are no new or partial Ethereum blocks to consume.
func (m *MsgIndex) NoEthereumSyncing() bool {
	return !m.IsFullEthereumSyncing() && !m.IsPartialEthereumSyncing()
}
