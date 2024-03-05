package types

// isEqualStringSlices compares two slices of strings for equality
func isEqualStringSlices(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false // Slices of different lengths cannot be equal
	}

	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false // Found unequal elements
		}
	}

	return true // All elements are equal
}

// Equal compares two EthEventsTx structs for equality based on their EventsData
func (m *EthEventsTx) Equal(e *EthEventsTx) bool {
	return isEqualStringSlices(m.EventsData, e.EventsData)
}
