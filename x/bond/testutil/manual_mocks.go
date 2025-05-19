package testutil

// MockAddressCodec is a mock implementation of address.Codec for testing purposes.
// This mock is used when we need a simple address codec that always returns a valid address.
type MockAddressCodec struct{}

// StringToBytes implements address.Codec interface.
// Returns a dummy address for testing.
func (m MockAddressCodec) StringToBytes(s string) ([]byte, error) {
	// Return a dummy address for testing
	return []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, nil
}

// BytesToString implements address.Codec interface.
// Not used in current tests, but required by the interface.
func (m MockAddressCodec) BytesToString(b []byte) (string, error) {
	return "", nil
}
