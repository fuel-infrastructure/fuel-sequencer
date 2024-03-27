package testutil

import (
	"encoding/hex"
	"fmt"
	"strings"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

func MustHexDecodeString(s string) []byte {
	// Check and remove the "0x" prefix if present
	if strings.HasPrefix(s, "0x") {
		s = s[2:]
	}

	decoded, err := hex.DecodeString(s)
	if err != nil {
		panic(fmt.Sprintf("MustDecodeString: invalid input %s", s))
	}

	return decoded
}

func MustGetSidecarEventFromParsedEvent(parsedEvent sidecartypes.ParsedEvent) *sidecartypes.Event {
	// Marshal the data
	data, err := parsedEvent.Marshal()
	if err != nil {
		panic(err)
	}

	// Get event type
	var eventType string
	switch parsedEvent.(type) {
	case *sidecartypes.SendToSequencerEvent:
		eventType = sidecartypes.SendToSequencerEventName
	case *sidecartypes.AuthorizeEvent:
		eventType = sidecartypes.AuthorizeEventName
	default:
		panic("invalid event type")
	}

	event := &sidecartypes.Event{
		EventType: eventType,
		Data:      data,
	}

	return event
}
