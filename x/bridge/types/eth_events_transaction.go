package types

import sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"

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
	return isEqualEventSlices(m.Events, e.Events)
}

// ValidateBasic performs some sanity checks on EthEventsTx
func (m *EthEventsTx) ValidateBasic() error {
	return isValidEventSlice(m.Events)
}
