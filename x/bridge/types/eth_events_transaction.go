package types

import (
	"errors"
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

// Equal compares two EthEventsTx structs and two sets of event transactions for equality
func (m *EthEventsTx) Equal(e *EthEventsTx, eventTxs1 [][]byte, eventTxs2 [][]byte) error {
	// If both structs are nil then they are equal
	if m == nil && e == nil {
		return nil
	}

	if m == nil || e == nil {
		return fmt.Errorf("nil (%t) != (%t)", m == nil, e == nil)
	} else if m.Authority != e.Authority {
		return fmt.Errorf("authority (%s) != (%s)", m.Authority, e.Authority)
	} else if m.NumInjectedEvents != e.NumInjectedEvents {
		return fmt.Errorf("number of injected events (%d) != (%d)", m.NumInjectedEvents, e.NumInjectedEvents)
	} else if m.NewEthereumBlock != e.NewEthereumBlock {
		return fmt.Errorf("new Ethereum block (%t) != (%t)", m.NewEthereumBlock, e.NewEthereumBlock)
	} else if m.BlockNumber != e.BlockNumber {
		return fmt.Errorf("block number (%d) != (%d)", m.BlockNumber, e.BlockNumber)
	} else if !utils.IsEqualBytesSlices(eventTxs1, eventTxs2) {
		return fmt.Errorf("event transactions are not equal")
	}
	return nil
}

// ValidateBeforeProcessing performs some state-based checks on EthEventsTx before it is officially processed.
func (m *EthEventsTx) ValidateBeforeProcessing(lastBlockSynced, eventIndexOffset uint64) error {

	// If we have an offset, we expect at least one new event.
	//
	// The cases where we receive NO events are:
	//
	// - New Ethereum block with no events - but since the offset is non-zero, we know that the current Ethereum block
	//   being synced still has more events for us to consume, so this case is invalid.
	// - No new Ethereum block - but since the offset is non-zero, we know that the current Ethereum block exists, and
	//   we expect to receive the remaining events. The current block is a new Ethereum block, so this case is invalid.
	if eventIndexOffset > 0 && m.NumInjectedEvents == 0 {
		return fmt.Errorf(
			"expected at least 1 new event if offset is non-zero (%d), got EthEventsTx (%s)",
			eventIndexOffset, m,
		)
	}

	// BlockNumber must be LastEthereumBlockSynced+1 since otherwise we're getting data for an Ethereum block
	// that we've already fully processed, or we're getting data for an Ethereum block that is in the future.
	expectedBlockNumber := lastBlockSynced + 1
	if m.BlockNumber != expectedBlockNumber {
		return fmt.Errorf(
			"expected block number %d, got %d in EthEventsTx (%s)",
			expectedBlockNumber, m.BlockNumber, m,
		)
	}

	return nil
}

// NumberOfEventsWithMaxBytes calculates the number of events that can fit into the specified maxBytes. This considers
// the size of the EthEventsTx as raw tx bytes and iterates over as many events as can fit into the specified maxBytes.
func (m *EthEventsTx) NumberOfEventsWithMaxBytes(eventTxs [][]byte, maxBytes uint64) (n int, err error) {
	rawTxBytes, err := m.RawTxBytes()
	if err != nil {
		return 0, err
	}
	n += len(rawTxBytes)
	for i, eventTx := range eventTxs {
		toAdd := len(eventTx)
		if uint64(n+toAdd) > maxBytes {
			return i, nil
		}
		n += toAdd
	}
	return len(eventTxs), nil
}

// TrimEventsFromHead removes the first N events from the front of the list of events.
//
// An important check that it does is to ensure that if there are events, these cannot all be trimmed, otherwise the
// blockchain might get stuck injecting empty EthEventsTx forever. At least one event must be kept if there are events.
func (m *EthEventsTx) TrimEventsFromHead(eventTxs [][]byte, numEventsToTrim uint64) ([][]byte, error) {

	if numEventsToTrim == 0 {

		return eventTxs, nil // trim nothing

	} else if m.NumInjectedEvents > 0 && m.NumInjectedEvents == numEventsToTrim {

		// If we trim all the events from the transaction this is a problem because if we retry
		// at the next block, we expect the same to happen, and we will never inject the events.
		return nil, fmt.Errorf("cannot trim all %d events", m.NumInjectedEvents)

	} else if numEventsToTrim > m.NumInjectedEvents {

		// If we try to trim more events than there are, something is wrong.
		return nil, fmt.Errorf(
			"insufficient no of events, expected at least %d got %d",
			numEventsToTrim, m.NumInjectedEvents,
		)

	}

	eventTxs = eventTxs[numEventsToTrim:]
	m.NumInjectedEvents = uint64(len(eventTxs))
	return eventTxs, nil
}

// KeepEventsFromHead keeps the first N events from the front of the list of events and trims the rest.
//
// An important check that it does is to ensure that if there are events, these cannot all be trimmed, otherwise the
// blockchain might get stuck injecting empty EthEventsTx forever. At least one event must be kept if there are events.
func (m *EthEventsTx) KeepEventsFromHead(
	eventTxs [][]byte, numEventsToKeep uint64,
) (newEventTxs [][]byte, trimmed uint64, err error) {

	if m.NumInjectedEvents == numEventsToKeep {

		return eventTxs, 0, nil // keep all

	} else if m.NumInjectedEvents > 0 && numEventsToKeep == 0 {

		// If we trim all the events from the transaction this is a problem because if we retry
		// at the next block, we expect the same to happen, and we will never inject the events.
		return nil, 0, fmt.Errorf("cannot trim all %d events", m.NumInjectedEvents)

	} else if m.NumInjectedEvents < numEventsToKeep {

		// If we try to trim more events than there are, something is wrong.
		return nil, 0, fmt.Errorf(
			"insufficient no of events, expected at least %d got %d",
			numEventsToKeep, m.NumInjectedEvents,
		)

	}

	eventTxs = eventTxs[:numEventsToKeep]
	trimmed = m.NumInjectedEvents - numEventsToKeep
	m.NumInjectedEvents = uint64(len(eventTxs))
	m.NewEthereumBlock = false

	// Note: changing NewEthereumBlock can affect the size of EthEventsTx. However, setting it to false will reduce
	// the size, not increase it, so there is no risk of exceeding the maxBytes as a result of setting it to false.

	return eventTxs, trimmed, nil
}

// RawTxBytes converts the message to a valid tx that can be injected into a block and produces a tx result.
func (m *EthEventsTx) RawTxBytes() ([]byte, error) {

	ethEventsTxAny, err := codectypes.NewAnyWithValue(m)
	if err != nil {
		return nil, err
	}

	ethEventsTxBz, err := utils.ValidRawTxBytesFromAnyMsgs([]*codectypes.Any{ethEventsTxAny})
	if err != nil {
		return nil, err
	}

	return ethEventsTxBz, nil
}

// FromSdkTx extracts EthEventsTx from an SDK transaction, which is expected to contain just EthEventsTx.
func (m *EthEventsTx) FromSdkTx(tx sdk.Tx) error {

	// EthEventsTx will contain only one message.
	msgs := tx.GetMsgs()
	if len(msgs) != 1 {
		return fmt.Errorf("expected 1 msg in EthEventsTx raw bytes, got %d", len(msgs))
	}

	// If the message is not a EthEventsTx continue with the other Ante decorators.
	msg := msgs[0]
	if sdk.MsgTypeURL(msg) != sdk.MsgTypeURL(&EthEventsTx{}) {
		return fmt.Errorf("expected msg type URL %s, got %s", sdk.MsgTypeURL(&EthEventsTx{}), sdk.MsgTypeURL(msg))
	}

	// If the message cannot be parsed into EthEventsTx, this is a problem.
	ethEventsTx, ok := msg.(*EthEventsTx)
	if !ok {
		return errors.New("could not parse message into EthEventsTx")
	}

	*m = *ethEventsTx
	return nil
}
