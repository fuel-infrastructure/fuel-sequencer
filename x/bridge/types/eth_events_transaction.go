package types

// isEqualStringSlices compares two slices of strings for equality
func isEqualStringSlices(slice1, slice2 []string) bool {
	// Slices of different lengths cannot be equal
	if len(slice1) != len(slice2) {
		return false
	}

	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false // Found unequal elements
		}
	}

	return true // All elements are equal
}

// TODO: Do equality and validatebasic for EthEventsTx

// Equal compares two EthEventsTx structs for equality
func (m *EthEventsTx) Equal(e *EthEventsTx) bool {
	return isEqualStringSlices(m.EventsData, e.EventsData)
}
