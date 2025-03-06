package utils_test

import (
	"testing"

	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

func TestIsEqualBytesSlices(t *testing.T) {

	bytes1 := [][]byte{[]byte("some bytes 1")}
	bytes2 := [][]byte{[]byte("some bytes 2")}

	bytes1And2 := append(bytes1, bytes2...)
	bytes2And1 := append(bytes2, bytes1...)

	testCases := []struct {
		name        string
		bytesSlice1 [][]byte
		bytesSlice2 [][]byte
		expectEqual bool
	}{
		{
			name:        "Equal bytes slices are considered equal",
			bytesSlice1: bytes1And2,
			bytesSlice2: bytes1And2,
			expectEqual: true,
		},
		{
			name:        "Unequal bytes slices are considered unequal",
			bytesSlice1: bytes1And2,
			bytesSlice2: bytes2And1,
			expectEqual: false,
		},
		{
			name:        "paddedBytes slices of different length are considered unequal",
			bytesSlice1: bytes1And2,
			bytesSlice2: bytes1,
			expectEqual: false,
		},
		{
			name:        "Nil bytes are considered equal",
			bytesSlice1: nil,
			bytesSlice2: nil,
			expectEqual: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			utils.IsEqualBytesSlices(tc.bytesSlice1, tc.bytesSlice2)

		})
	}
}
