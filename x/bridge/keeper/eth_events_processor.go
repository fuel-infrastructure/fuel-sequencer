package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

// ProcessEthereumEvents processes the Ethereum events injected at lastEthereumBlockSynced
func (k Keeper) ProcessEthereumEvents(ctx sdk.Context) {
	lastEthereumBlockSynced := k.MustGetLastEthereumBlockSynced(ctx)

	// Get EthEventsTx at lastEthereumBlockSynced
	ethEventsTx, found := k.GetEthEventsTx(ctx, lastEthereumBlockSynced.Uint64())
	if !found {
		// If EthEventsTx is not found then either the block has already been processed or no new events where generated
		return
	}

	for _, event := range ethEventsTx.Events {
		// Unmarshal event sent by sidecar to a parsedEvent
		parsedEvent, err := event.UnmarshalParsedEvent()
		if err != nil {
			// If an event cannot be unmarshalled to ParsedEvent ignore it and move on to the next event as there
			// might be something very suspicious with the event
			k.Logger().Error("Bridge EndBlock: could not unmarshal to ParsedEvent", "event", event.String(), "err", err)
			continue
		}

		// Process event based on its type. If an error occurs while processing an event we will not apply any state
		// changes and move on to the next event.
		switch pe := parsedEvent.(type) {
		case *sidecartypes.SendToSequencerEvent:
			err = utils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
				err = k.ProcessSendToSequencerEvent(ctx, pe)
				return err
			})
			if err != nil {
				k.Logger().Error(
					"Bridge EndBlock: failed to process SendToSequencerEvent",
					"event", pe.String(),
					"err", err,
				)
				continue
			}
		case *sidecartypes.AuthorizeEvent:
			err = utils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
				err = k.ProcessAuthorizeEvent(ctx, pe)
				return err
			})
			if err != nil {
				k.Logger().Error("Bridge EndBlock: failed to process AuthorizeEvent", "event", pe.String(), "err", err)
				continue
			}
		default:
			// If an event type is unrecognized ignore the event and move on to the next as there might be something
			// very suspicious with the event
			k.Logger().Error(fmt.Sprintf("Bridge EndBlock: unexpected ParsedEvent type %T", pe))
			continue
		}
	}

	k.RemoveEthEventsTx(ctx, lastEthereumBlockSynced.Uint64())

	return
}

// ProcessSendToSequencerEvent attempts to process a SendToSequencerEvent
func (k Keeper) ProcessSendToSequencerEvent(ctx sdk.Context, event *sidecartypes.SendToSequencerEvent) error {

	return nil
}

// ProcessAuthorizeEvent attempts to process an AuthorizeEvent
func (k Keeper) ProcessAuthorizeEvent(ctx sdk.Context, event *sidecartypes.AuthorizeEvent) error {
	// TODO: Authorize Messages, all pass or none at all, return error in this case.
	// TODO: Do I need to save cache ctx at this stage? Probably yes because we are saving to cache not to state. Something to check.
	// TODO: ICA like logic inside of switch statement.
	// TODO: Double context is important so we control when an error is written. Write error only if not all messages
	//     : can be processed, I think we don't need to cache double context because the outside is enough.
	return nil
}
