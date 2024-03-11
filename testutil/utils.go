package testutil

import (
	"encoding/hex"
	"fmt"
)

func MustHexDecodeString(s string) []byte {
	decoded, err := hex.DecodeString(s)
	if err != nil {
		panic(fmt.Sprintf("MustDecodeString: invalid input %s", s))
	}
	return decoded
}
