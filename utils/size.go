package utils

// NumberOfBytes returns the total amount of bytes that a [][]byte slice occupies.
func NumberOfBytes(slice [][]byte) (numberOfBytes uint64) {
	for _, element := range slice {
		numberOfBytes += uint64(len(element))
	}

	return numberOfBytes
}
