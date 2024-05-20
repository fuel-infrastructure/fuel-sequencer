package utils

import "bytes"

// IsEqualBytesSlices compares two slices of []byte for equality.
func IsEqualBytesSlices(slice1, slice2 [][]byte) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i := range slice1 {
		if !bytes.Equal(slice1[i], slice2[i]) {
			return false
		}
	}

	return true
}
