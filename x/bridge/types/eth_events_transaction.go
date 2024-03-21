package types

import (
	"errors"

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

	return m.AdvanceSequencer == e.AdvanceSequencer &&
		m.NewEthereumBlock == e.NewEthereumBlock &&
		m.BlockNumber.Equal(e.BlockNumber) &&
		equalEventSlices, err
}

// ValidateBasic performs some sanity checks on EthEventsTx
func (m *EthEventsTx) ValidateBasic() error {
	// Error if the receiver is nil
	if m == nil {
		return errors.New("EthEventsTx is nil")
	}

	return isValidEventSlice(m.Events)
}
