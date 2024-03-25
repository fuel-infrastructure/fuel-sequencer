package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

// ProcessEthereumEvents processes the Ethereum events injected at a specific height
func (k Keeper) ProcessEthereumEvents(ctx context.Context, blockHeight math.Int) {
	// Get EthEventsTx at specified block height
	ethEventsTx, found := k.GetEthEventsTx(ctx, blockHeight.Uint64())
	if !found {
		// If EthEventsTx is not found then either the block has already been processed or no new events where generated
		return
	}

	// TODO: Attempt to process event and ignore it if it is not succesful, don't return error in this case
	for _, event := range ethEventsTx.Events {
		parsedEvent, err := event.UnmarshalParsedEvent()
		// TODO: Log error and move on to the next event if it cannot be unmarshalled

		// Perform operation based on trade order status. Trade order is passed by reference
		// and will be saved at the end of this function with any modifications.
		switch parsedEvent.(type) {
		case *sidecartypes.SendToSequencerEvent:
			// TODO: ApplyFuncIfNoErr
			err = k.ProcessTradeAtInit(ctx, tradeOrder)
		case *sidecartypes.AuthorizeEvent:
			// TODO: ApplyFuncIfNoErr
			var tradeOrderCompleted bool
			tradeOrderCompleted, err = k.ProcessTradeAtIBCTransferOutgoing(ctx, tradeOrder)
			if tradeOrderCompleted {
				updateTradeOrder = false // Trading is done, so no need to update trade order
			}
		default:
			panic(fmt.Sprintf("unexpected trade order status %s for trade order %d", tradeOrder.Status, tradeOrder.Id))
		}
	}

	k.RemoveEthEventsTx(ctx, blockHeight.Uint64())

	return
}

// ProcessSendToSequencerEvent attempts to process a SendToSequencerEvent
func (k Keeper) ProcessSendToSequencerEvent(ctx context.Context, event sidecartypes.SendToSequencerEvent) error {

	return nil
}

// ProcessAuthorizeEvent attempts to process an AuthorizeEvent
func (k Keeper) ProcessAuthorizeEvent(ctx context.Context, event sidecartypes.AuthorizeEvent) error {
	// TODO: Authorize Messages, all pass or none at all, return error in this case.
	// TODO: Do I need to save cache ctx at this stage? Probably yes because we are saving to cache not to state. Something to check.
	// TODO: ICA like logic inside of switch statement.
	// TODO: Double context is important so we control when an error is written. Write error only if not all messages
	//     : can be processed, I think we don't need to cache double context because the outside is enough.
	return nil
}
