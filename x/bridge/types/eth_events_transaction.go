package types

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

// isEqualEventSlices compares two slices of Ethereum events for equality
func isEqualEventSlices(slice1, slice2 []*sidecartypes.Event) (bool, error) {
	if len(slice1) != len(slice2) {
		// Slices of different lengths cannot be equal
		return false, nil
	}

	for i := range slice1 {
		equalElements, err := slice1[i].Equal(slice2[i])
		if err != nil {
			// Return error if equality failed
			return false, err
		}

		if !equalElements {
			// Found unequal elements
			return false, nil
		}
	}

	// All elements are equal
	return true, nil
}

// isValidEventSlice performs some sanity checks on the list of Ethereum events
func isValidEventSlice(slice []*sidecartypes.Event) error {
	for _, event := range slice {
		err := event.ValidateBasic()
		if err != nil {
			// Return error if verification failed.
			return err
		}
	}

	// All elements pass validation
	return nil
}

// Equal compares two EthEventsTx structs for equality
func (m *EthEventsTx) Equal(e *EthEventsTx) (bool, error) {
	// If both structs are nil then they are equal
	if m == nil && e == nil {
		return true, nil
	}

	// If one of them only is nil then they are not equal
	if m == nil || e == nil {
		return false, nil
	}

	// Check if the events are equal
	equalEventSlices, err := isEqualEventSlices(m.Events, e.Events)
	if err != nil {
		return false, err
	}

	return m.AdvanceSequencer == e.AdvanceSequencer &&
		m.NewEthereumBlock == e.NewEthereumBlock &&
		m.BlockNumber.Equal(e.BlockNumber) &&
		equalEventSlices, nil
}

// ValidateBasic performs some sanity checks on EthEventsTx
func (m *EthEventsTx) ValidateBasic() error {
	// Error if the receiver is nil
	if m == nil {
		return errors.New("EthEventsTx is nil")
	}

	return isValidEventSlice(m.Events)
}

// ValidateBeforeProcessing performs some state-based checks on EthEventsTx before it is officially processed.
func (m *EthEventsTx) ValidateBeforeProcessing(lastBlockSynced, eventIndexOffset sdkmath.Int) error {

	// We cannot process an EthEventsTx without advancing the sequencer.
	if !m.AdvanceSequencer {
		return fmt.Errorf("expected AdvanceSequencer to be true, got false in EthEventsTx (%s)", m)
	}

	// If we have an offset, we expect at least one new event.
	//
	// The cases where we receive NO events are:
	// - New Ethereum block with no events ...but we know that the current Ethereum block still has more events.
	// - No new Ethereum block ...but we know that the current Ethereum block exists and still has more events.
	// - An error occurred and AdvanceSequencer is false ...but we know that this is no possible from the check above.
	if eventIndexOffset.IsPositive() && len(m.Events) == 0 {
		return fmt.Errorf(
			"expected at least 1 new event if offset is non-zero (%s), got EthEventsTx (%s)",
			eventIndexOffset, m,
		)
	}

	// BlockNumber must be LastEthereumBlockSynced+1 since otherwise we're getting data for an Ethereum block
	// that we've already fully processed, or we're getting data for an Ethereum block that is in the future.
	expectedBlockNumber := lastBlockSynced.Add(sdkmath.OneInt())
	if !m.BlockNumber.Equal(expectedBlockNumber) {
		return fmt.Errorf(
			"expected block number %s, got %s in EthEventsTx (%s)",
			expectedBlockNumber, m.BlockNumber, m,
		)
	}

	return nil
}

// NumberOfEventsWithMaxBytes calculates the number of event that can fit into the specified maxBytes. This closely
// resembles the EthEventsTx Size function but only iterates over as many events as can fit into the specified maxBytes.
//
// It is very important to update this function if the EthEventsTx Size function gets updated, otherwise we might be
// overestimating or underestimating the size of EthEventsTx and inject a suboptimal number of events.
func (m *EthEventsTx) NumberOfEventsWithMaxBytes(maxBytes uint64) (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.AdvanceSequencer {
		n += 2
	}
	if m.NewEthereumBlock {
		n += 2
	}
	l = m.BlockNumber.Size()
	n += 1 + l + sovEthEventsTransaction(uint64(l))
	if len(m.Events) > 0 {
		for i, e := range m.Events {
			l = e.Size()
			toAdd := 1 + l + sovEthEventsTransaction(uint64(l))
			if uint64(n+toAdd) > maxBytes {
				return i
			}
			n += toAdd
		}
	}
	return len(m.Events)
}

// TrimEventsFromHead removes the first N events from the front of the list of events.
//
// An important check that it does is to ensure that if there are events, these cannot all be trimmed, otherwise the
// blockchain might get stuck injecting empty EthEventsTx forever. At least one event must be kept if there are events.
func (m *EthEventsTx) TrimEventsFromHead(numEventsToTrim uint64) error {
	numEventsInTx := uint64(len(m.Events))

	if numEventsInTx > 0 && numEventsInTx == numEventsToTrim {
		return fmt.Errorf("cannot trim all %d events from EthEventsTx %s", numEventsInTx, m)
	}

	if numEventsToTrim == 0 {
		return nil // trim nothing
	} else if numEventsToTrim > numEventsInTx {
		return fmt.Errorf("insufficient no of events, expected at least %d got %d", numEventsToTrim, numEventsInTx)
	} else {
		m.Events = m.Events[numEventsToTrim:]
	}

	return nil
}

// KeepEventsFromHead keeps the first N events from the front of the list of events and trims the rest.
//
// An important check that it does is to ensure that if there are events, these cannot all be trimmed, otherwise the
// blockchain might get stuck injecting empty EthEventsTx forever. At least one event must be kept if there are events.
func (m *EthEventsTx) KeepEventsFromHead(numEventsToKeep uint64) (trimmed uint64, err error) {
	numEventsInTx := uint64(len(m.Events))

	// If we cannot fit any events into the block this is a problem because if we retry at
	// the next block, we expect the same to happen, and we will never inject the events.
	if numEventsInTx > 0 && numEventsToKeep == 0 {
		return 0, fmt.Errorf("cannot trim all %d events from EthEventsTx %s", numEventsInTx, m)
	}

	if numEventsInTx == numEventsToKeep {
		return 0, nil // keep all
	} else if numEventsInTx < numEventsToKeep {
		return 0, fmt.Errorf("insufficient no of events, expected at least %d got %d", numEventsToKeep, numEventsInTx)
	} else {
		m.Events = m.Events[:numEventsToKeep]
		m.NewEthereumBlock = false

		// Note: changing NewEthereumBlock can affect the size of EthEventsTx. However, setting it to false will reduce
		// the size, not increase it, so there is no risk of exceeding the maxBytes as a result of setting it to false.

		return numEventsInTx - numEventsToKeep, nil
	}
}
