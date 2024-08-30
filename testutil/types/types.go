package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// TestMsgIndexWithEvents is a convenient combination of a MsgIndex and a slice of events.
type TestMsgIndexWithEvents struct {
	*bridgetypes.MsgIndex
	Events []*sidecartypes.Event
}

// MustGetDepositMsgFromDepositEvent extracts a single deposit message from an Ethereum event and panics otherwise.
func MustGetDepositMsgFromDepositEvent(
	cdc codec.BinaryCodec, authority string, depositEvent *sidecartypes.Event,
) *bridgetypes.MsgDepositFromEthereum {

	msgs, err := depositEvent.Messages(cdc, authority)
	if err != nil {
		panic(err)
	} else if len(msgs) > 1 {
		panic("expected only one message from test event")
	}

	var msg sdk.Msg
	err = cdc.UnpackAny(msgs[0], &msg)
	if err != nil {
		panic(err)
	}

	depositMsg, ok := msg.(*bridgetypes.MsgDepositFromEthereum)
	if !ok {
		panic("could not infer deposit message")
	}

	return depositMsg
}

// MustGetEventTxsFromEvents gets raw tx bytes from all the passed events, and panics otherwise.
func MustGetEventTxsFromEvents(
	cdc codec.BinaryCodec, authority string, events []*sidecartypes.Event,
) (allRawTxBytes [][]byte) {

	for _, event := range events {
		rawTxBytes, err := event.RawTxBytes(cdc, authority)
		if err != nil {
			panic(err)
		}

		allRawTxBytes = append(allRawTxBytes, rawTxBytes)
	}
	return
}

// MustGetSizeFromEvents gets the size of the raw tx bytes from all the passed events, and panics otherwise.
func MustGetSizeFromEvents(cdc codec.BinaryCodec, authority string, events []*sidecartypes.Event) int {
	return int(utils.TxsSize(MustGetEventTxsFromEvents(cdc, authority, events)))
}
