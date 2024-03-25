package keeper

import (
	"context"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

// ProcessBridgeMessages processes the messages that we have in store
func (k Keeper) ProcessBridgeMessages(ctx context.Context) {

	// Firstly retrieve the blockHeight we have last processed.
	lastEthereumBlockSynced := k.MustGetLastEthereumBlockSynced(ctx)

	// Secondly retireve the events and the last block height processed.
	events, ok := k.GetEthEventsTx(ctx, lastEthereumBlockSynced.Uint64())
	if !ok {
		// If no events arre found that this block height then we can stop here
		return
	}

	// Otherwise we being processing these events according to event type
	for _, event := range events.Events {
		switch event.EventType {
		case sidecartypes.AuthorizeEventName:
			// TODO process auth event
		case sidecartypes.SendToSequencerEventName:
			// TODO process send event
		default:
			// TODO ERROR HERE
		}
	}
}
