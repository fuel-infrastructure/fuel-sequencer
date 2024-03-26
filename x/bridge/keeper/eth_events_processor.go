package keeper

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
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
func (k Keeper) ProcessSendToSequencerEvent(_ sdk.Context, _ *sidecartypes.SendToSequencerEvent) error {

	return nil
}

// ProcessAuthorizeEvent attempts to process an AuthorizeEvent
func (k Keeper) ProcessAuthorizeEvent(ctx sdk.Context, event *sidecartypes.AuthorizeEvent) error {

	// Deserialize AuthorizeEvent.Message into an array of sdk.Msg
	msgs, err := k.DeserializeAuthorizeTx(k.cdc, event)
	if err != nil {
		return types.ErrCouldNotDeserializeAuthorizeTx.Wrapf("%v", err)
	}

	if err = k.authenticateTx(ctx, event.From, msgs); err != nil {
		return types.ErrCouldNotAuthenticateTx.Wrapf("%v", err)
	}

	// Execute every deserialized msg. If one of the messages errors during execution we will revert the state. i.e.
	// either all messages get executed successfully or none at all.
	txMsgData := &sdk.TxMsgData{
		MsgResponses: make([]*codectypes.Any, len(msgs)),
	}
	err = utils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		for index, msg := range msgs {

			// Confirm that the message passes the necessary stateless checks
			if m, ok := msg.(sdk.HasValidateBasic); ok {
				if err := m.ValidateBasic(); err != nil {
					return types.ErrCouldNotValidateMsg.Wrapf("msg: %s err: %v", msg.String(), err)
				}
			}

			// Execute message and store the response
			msgResponse, err := k.executeMsg(ctx, msg)
			if err != nil {
				return types.ErrCouldNotExecuteMsg.Wrapf("msg: %s err: %v", msg.String(), err)
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

// authenticateTx ensures that the msgs signer is the mapped Sequencer address of the sender
func (k Keeper) authenticateTx(ctx sdk.Context, sender string, msgs []sdk.Msg) error {

	// Generate the Sequencer address from the Ethereum address
	mappedSequencerAddr, err := types.GenerateSequencerAddressFromEthereumAddress(sender)
	if err != nil {
		return fmt.Errorf("could not generate Sequencer address from Ethereum address: %w", err)
	}

	messagesAllowed := k.GetParams(ctx).AuthorizeMessagesAllowed
	for _, msg := range msgs {

		// Check that the message is authorized
		if !AuthorizedMessage(messagesAllowed, msg) {
			return fmt.Errorf("message %s not authorized on Sequencer", sdk.MsgTypeURL(msg))
		}

		// Obtain the message signers using the proto signer annotations
		protoCodec, ok := k.cdc.(*codec.ProtoCodec)
		if !ok {
			return errors.New(
				"codec is not supported: only the ProtoCodec may be used for receiving messages on the Sequencer",
			)
		}
		signers, _, err := protoCodec.GetMsgV1Signers(msg)
		if err != nil {
			return fmt.Errorf("failed to obtain message signers for message type %s: %w", sdk.MsgTypeURL(msg), err)
		}

		for _, signer := range signers {

			// Make sure that the message signer is equivalent to the mapped Sequencer address of the sender on Ethereum
			if mappedSequencerAddr.String() != sdk.AccAddress(signer).String() {
				return fmt.Errorf(
					"unexpected signer address: expected %s, got %s",
					mappedSequencerAddr.String(),
					sdk.AccAddress(signer).String(),
				)
			}
		}
	}

	return nil
}

// ExecuteMsg attempts to execute an authorized message originating from Ethereum
func (k Keeper) executeMsg(_ sdk.Context, _ sdk.Msg) (*codectypes.Any, error) {
	return nil, nil
}
