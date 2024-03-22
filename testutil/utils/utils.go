package utils

import "encoding/hex"

// MockTopicIDHex generates a mock 32-byte hash for testing, represented as a hexadecimal string, based on an input number.
func MockTopicIDHex(num int) []byte {
	// Convert the input number to a hexadecimal string, ensuring it's 64 characters long for a 32-byte hash.
	hexStr := hex.EncodeToString([]byte{byte(num)})
	// Pad the string to ensure it's 32 bytes long when decoded.
	for len(hexStr) < 64 {
		hexStr += "0"
	}
	b, _ := hex.DecodeString(hexStr)
	return b
}
