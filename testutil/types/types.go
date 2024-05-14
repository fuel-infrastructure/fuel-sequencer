package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// TestEthEventsTxWithEvents is a convenient combination of an EthEventsTx and a slice of events.
type TestEthEventsTxWithEvents struct {
	*bridgetypes.EthEventsTx
	Events []*sidecartypes.Event
}

// MustGetDepositMsgFromDepositEvent extracts a single deposit message from an Ethereum event and panics otherwise.
func MustGetDepositMsgFromDepositEvent(
	cdc codec.BinaryCodec, depositEvent *sidecartypes.Event,
) *bridgetypes.MsgDepositFromEthereum {

	msgs, err := depositEvent.Messages(cdc)
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

// MustGetRawTxBytesFromEvents gets raw tx bytes from all the passed events, and panics otherwise.
func MustGetRawTxBytesFromEvents(cdc codec.BinaryCodec, events []*sidecartypes.Event) (allRawTxBytes [][]byte) {

	for _, event := range events {
		rawTxBytes, err := event.RawTxBytes(cdc)
		if err != nil {
			panic(err)
		}

		allRawTxBytes = append(allRawTxBytes, rawTxBytes)
	}
	return
}

// MustGetSizeFromEvents gets the size of the raw tx bytes from all the passed events, and panics otherwise.
func MustGetSizeFromEvents(cdc codec.BinaryCodec, events []*sidecartypes.Event) (size int) {

	for _, event := range MustGetRawTxBytesFromEvents(cdc, events) {
		size += len(event)
	}
	return
}
