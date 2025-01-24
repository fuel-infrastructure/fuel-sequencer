package testutil

import (
	"encoding/hex"
	"fmt"
	"strings"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

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

func MustHexDecodeString(s string) []byte {
	// Remove the "0x" prefix if present
	s = strings.TrimPrefix(s, "0x")

	decoded, err := hex.DecodeString(s)
	if err != nil {
		panic(fmt.Sprintf("MustDecodeString: invalid input %s", s))
	}

	return decoded
}

func MustGetSidecarEventFromParsedEvent(
	parsedEvent sidecartypes.ParsedEvent, ethereumProxyContractAddress string,
) *sidecartypes.Event {
	// Marshal the data
	data, err := parsedEvent.Marshal()
	if err != nil {
		panic(err)
	}

	// Get event type
	var eventType string
	switch parsedEvent.(type) {
	case *sidecartypes.DepositEvent:
		eventType = sidecartypes.DepositEventName
	case *sidecartypes.AuthorizeEvent:
		eventType = sidecartypes.AuthorizeEventName
	default:
		panic("invalid event type")
	}

	event := &sidecartypes.Event{
		EventType:       eventType,
		ContractAddress: ethereumProxyContractAddress,
		Data:            data,
	}

	return event
}
