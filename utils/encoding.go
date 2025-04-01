package utils

import (
	"encoding/hex"
	"fmt"
	"strconv"
)

// To32PaddedHexBytes takes a number and returns its hex representation padded to 32 bytes.
// Used to mimic the result of `abi.encode(number)` in Ethereum.
// Ref: https://docs.soliditylang.org/en/develop/abi-spec.html
func To32PaddedHexBytes(number uint64) ([]byte, error) {
	hexRepresentation := strconv.FormatUint(number, 16)
	// Make sure hex representation has even length.
	// The `strconv.FormatUint` can return odd length hex encodings.
	// For example, `strconv.FormatUint(10, 16)` returns `a`.
	// Thus, we need to pad it.
	if len(hexRepresentation)%2 == 1 {
		hexRepresentation = "0" + hexRepresentation
	}
	hexBytes, hexErr := hex.DecodeString(hexRepresentation)
	if hexErr != nil {
		return nil, hexErr
	}
	paddedBytes, padErr := padBytes(hexBytes, 32)
	if padErr != nil {
		return nil, padErr
	}
	return paddedBytes, nil
}

// padBytes Pad bytes to given length.
func padBytes(byt []byte, length int) ([]byte, error) {
	l := len(byt)
	if l > length {
		return nil, fmt.Errorf(
			"cannot pad bytes because length of bytes array: %d is greater than given length: %d",
			l,
			length,
		)
	}
	if l == length {
		return byt, nil
	}
	tmp := make([]byte, length)
	copy(tmp[length-l:], byt)
	return tmp, nil
}
