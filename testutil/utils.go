package testutil

import (
	"encoding/hex"
	"fmt"
	"strings"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

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
		eventType = sidecartypes.MockDepositEventName
	case *sidecartypes.AuthorizeEvent:
		eventType = sidecartypes.MockAuthorizeEventName
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
