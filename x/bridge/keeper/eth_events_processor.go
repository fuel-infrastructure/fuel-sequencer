package keeper

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
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

	// Deserialize AuthorizeEvent.Message into an array of sdk.Msg
	msgs, err := k.DeserializeAuthorizeTx(k.cdc, event)
	if err != nil {
		return types.ErrCouldNotDeserializeAuthorizeTx.Wrapf("%w", err)
	}

	// TODO: Authenticate Tx here and return error if authentication fails

	// Execute every deserialized msg. If one of the messages errors during execution we will revert the state. i.e.
	// either all messages get executed successfully or none at all.
	txMsgData := &sdk.TxMsgData{
		MsgResponses: make([]*codectypes.Any, len(msgs)),
	}
	err = utils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {

		// TODO: Perform validate basic here and return error if one of the msgs fail validation

		for index, msg := range msgs {
			msgResponse, err := k.executeMsg(ctx, msg)
			if err != nil {
				return types.ErrCouldExecuteMsg.Wrapf("msg: %s err: %w", msg.String(), err)
			}
			txMsgData.MsgResponses[index] = msgResponse
		}

		return nil
	})
	if err != nil {
		return err
	}

	err = ctx.EventManager().EmitTypedEvent(&types.EventAuthorizedTxExecuted{MsgResponses: txMsgData.MsgResponses})
	if err != nil {
		return err
	}

	return nil
}

// ExecuteMsg attempts to execute an authorized message originating from Ethereum
func (k Keeper) executeMsg(ctx sdk.Context, msg sdk.Msg) (*codectypes.Any, error) {
	return nil, nil
}
