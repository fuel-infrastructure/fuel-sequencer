package keeper

import (
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// ProcessEthereumEvents processes the Ethereum events injected at lastEthereumBlockSynced
func (k Keeper) ProcessEthereumEvents(ctx sdk.Context) {

	// Firstly retrieve the blockHeight we have last processed.
	lastEthereumBlockSynced := k.MustGetLastEthereumBlockSynced(ctx)

	// Secondly retrieve the events from the last block height processed.
	ethEventsTx, ok := k.GetEthEventsTx(ctx, lastEthereumBlockSynced.Uint64())
	if !ok {
		// If no events are found at this block height then we can stop here.
		return
	}

	// Get the params as they are needed for the denom.
	params := k.GetParams(ctx)
	supplyDelta := k.MustGetSupplyDeltaInfo(ctx)

	// Otherwise we begin processing these events according to the event type.
	for _, event := range ethEventsTx.Events {

		// Unmarshal the parsed event if possible.
		parsedEvent, err := event.UnmarshalParsedEvent()
		if err != nil {
			// If an event cannot be unmarshalled to ParsedEvent ignore it and move on to the next event as there
			// might be something very suspicious with the event
			k.Logger().Error("Bridge EndBlock: could not unmarshal to ParsedEvent", "event", event.String(), "err", err)
			continue
		}

		// Process event based on its type.
		switch pe := parsedEvent.(type) {
		case *sidecartypes.SendToSequencerEvent:
			// This doesn't error, so unless a panic occurs we will always be able to continue to the next event if some
			// issue occurs
			k.processSendToSequencerEvent(ctx, pe, params, &supplyDelta)
		case *sidecartypes.AuthorizeEvent:
			// If an error occurs while processing an Authorize event we will move on to the next event without applying
			// any state changes.
			err = utils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
				return k.processAuthorizeEvent(ctx, pe, &params)
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
}

// processSendToSequencerEvent processes the send to sequencer events queried from the sidecar.
// Deposit message cannot fail. If a failure can occur, we mint the tokens anyway
// but we store them in the governance address. The only time it can fail is if
// we cannot parse the `Amount` of tokens as we won't know how many tokens have been processed.
func (k Keeper) processSendToSequencerEvent(
	ctx sdk.Context,
	sendEvent *sidecartypes.SendToSequencerEvent,
	params types.Params,
	supplyDeltaInfo *types.SupplyDeltaInfo,
) {

	// TODO: Check that the contract address matches the stored proxy contract address.

	// Parse the data accordingly.
	amount, success := sdkmath.NewIntFromString(sendEvent.Amount)
	if !success {
		// Panic if we've failed to unmarshal an amount, as something has gone wrong in the entire chain process.
		k.Logger().Error("Bridge EndBlock: could not unmarshal amount to int from string")
		panic("Bridge EndBlock: could not unmarshal amount to int from string")
	}
	tokenToMint := sdk.NewCoin(params.BridgeDenom, amount)
	tokensToMint := sdk.NewCoins(tokenToMint)

	// Check that the Duration can be converted from a string to sdk.Int
	eventDuration, success := sdkmath.NewIntFromString(sendEvent.Duration)
	if !success {
		k.Logger().Error("Bridge EndBlock: failed to process send to sequencer duration from string")
		k.mintToGovernanceAddress(ctx, tokenToMint, supplyDeltaInfo)
		return
	}

	// Convert the duration in seconds to a vesting duration
	vesting := time.Duration(eventDuration.Int64() * 1e9)

	// Check that From is a valid hex address
	if !common.IsHexAddress(sendEvent.From) {
		k.Logger().Error(
			"Bridge EndBlock: from address is not a valid hex address - minting to governance address instead",
			"event", sendEvent,
		)
		k.mintToGovernanceAddress(ctx, tokenToMint, supplyDeltaInfo)
		return
	}

	// sequencerAddr to be determined based on data provided in event.
	var sequencerAddr sdk.AccAddress
	var err error

	// If a `To` address was not specified send tokens to the address mapped 1-to-1 fom the `From` Ethereum Address.
	if len(strings.TrimSpace(sendEvent.To)) == 0 {
		sequencerAddr, err = k.generateSequencerAccountFromEthereumDeposit(ctx, sendEvent.From, vesting, tokensToMint)
		if err != nil {
			k.Logger().Error(
				"Bridge EndBlock: failed to generate sequencer account from ethereum address - minting to gov address",
				"event", sendEvent, "err", err,
			)
			k.mintToGovernanceAddress(ctx, tokenToMint, supplyDeltaInfo)
			return
		}
	} else {

		// Otherwise process the To from a string to an AccAddress type.
		sequencerAddr, err = sdk.AccAddressFromBech32(sendEvent.To)
		if err != nil {
			k.Logger().Error(
				"Bridge EndBlock: to is not a valid Bech32 address - minting to gov address",
				"event", sendEvent, "err", err,
			)
			k.mintToGovernanceAddress(ctx, tokenToMint, supplyDeltaInfo)
			return
		}
	}

	// Otherwise mint and send the coins to the specified user.
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, tokensToMint)
	if err != nil {
		k.Logger().Error("Bridge EndBlock: failed to mint tokens to module", "err", err)
		panic(err)
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sequencerAddr, tokensToMint)
	if err != nil {
		k.Logger().Error("Bridge EndBlock: failed to send tokens from module to account", "err", err)
		panic(err)
	}

	// Apply negative offset to supply delta offset
	supplyDeltaInfo.Offset = supplyDeltaInfo.Offset.Sub(amount)

	// We have to save the supply delta here incase we panic at a later deposit.
	k.SetSupplyDeltaInfo(ctx, *supplyDeltaInfo)

	// Emit event once completed
	err = ctx.EventManager().EmitTypedEvent(&types.EventSendToSequencerEventProcessed{
		From:     sendEvent.From,
		To:       sequencerAddr.String(),
		Amount:   tokenToMint,
		Duration: vesting.String(),
	})
	if err != nil {
		k.Logger().Error("Bridge EndBlock: failed to emit event send to sequencer", "err", err)
	}

	k.Logger().Debug(
		"Bridge EndBlock: minted bridge tokens to account",
		"amount", tokenToMint.Amount, "address", sequencerAddr,
	)
}

// mintToGovernanceAddress mints to the governance address in case of an error in normal processing.
func (k Keeper) mintToGovernanceAddress(ctx sdk.Context, tokenToMint sdk.Coin, supplyDeltaInfo *types.SupplyDeltaInfo) {

	// tokensToMint is the new coins that will be minted
	tokensToMint := sdk.NewCoins(tokenToMint)

	// Mint bridged tokens to governance module address.
	err := k.bankKeeper.MintCoins(ctx, govtypes.ModuleName, tokensToMint)
	if err != nil {
		panic(fmt.Errorf("failed to mint bridge tokens to gov module account err: %s", err))
	}

	// Update supply delta to reflect the minting to the governance address.
	supplyDeltaInfo.Offset = supplyDeltaInfo.Offset.Sub(tokenToMint.Amount)

	// Save supply delta here after processing mints.
	k.SetSupplyDeltaInfo(ctx, *supplyDeltaInfo)

	// Emit event once completed
	err = ctx.EventManager().EmitTypedEvent(&types.EventSendToSequencerEventProcessed{
		From:   types.ModuleName,
		To:     govtypes.ModuleName,
		Amount: tokenToMint,
	})
	if err != nil {
		k.Logger().Error("Bridge EndBlock: failed to emit event send to sequencer", "err", err)
	}

	k.Logger().Warn("minted bridge tokens to governance address", "amount", tokenToMint.Amount)
}

// processAuthorizeEvent attempts to process an AuthorizeEvent by executing all of its messages
func (k Keeper) processAuthorizeEvent(ctx sdk.Context, event *sidecartypes.AuthorizeEvent, params *types.Params) error {
	// Deserialize AuthorizeEvent.Message into an array of sdk.Msg
	msgs, err := types.DeserializeAuthorizeTx(k.cdc, event)
	if err != nil {
		return fmt.Errorf("could not deserialize AuthorizeTx: %w", err)
	}

	// Check whether AuthorizeTx is authorized on the Sequencer
	if err = k.authenticateTx(event.From, msgs, params); err != nil {
		return fmt.Errorf("could not authenticate AuthorizeTx: %w", err)
	}

	// Execute every deserialized msg. If one of the messages errors during execution we will revert the state. i.e.
	// either all messages get executed successfully or none at all.
	err = utils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		for _, msg := range msgs {

			// Confirm that the message passes the necessary stateless checks
			if m, ok := msg.(sdk.HasValidateBasic); ok {
				if err := m.ValidateBasic(); err != nil {
					return fmt.Errorf("could not validate msg: msg %s, err: %w", msg.String(), err)
				}
			}

			// Execute message
			err := k.executeMsg(ctx, msg)
			if err != nil {
				return fmt.Errorf("could not execute msg: msg %s, err: %w", msg.String(), err)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// authenticateTx ensures that the msgs signer is the mapped Sequencer address of the sender
func (k Keeper) authenticateTx(sender string, msgs []sdk.Msg, params *types.Params) error {

	// Generate the Sequencer address from the Ethereum address
	mappedSequencerAddr, err := types.GenerateSequencerAddressFromEthereumAddress(sender)
	if err != nil {
		return types.ErrCouldNotGenerateSequencerAddress.Wrapf("%v", err)
	}

	for _, msg := range msgs {

		// Check that the message is authorized
		if !params.IsAuthorizedMessage(msg) {
			return types.ErrMsgNotAuthorizedOnSequencer.Wrapf("%s", sdk.MsgTypeURL(msg))
		}

		// Obtain the message signers using the proto signer annotations
		protoCodec, ok := k.cdc.(*codec.ProtoCodec)
		if !ok {
			return types.ErrCodecIsNotSupported.Wrap(types.ErrStrOnlyProtoCodecAllowed)
		}
		signers, _, err := protoCodec.GetMsgV1Signers(msg)
		if err != nil {
			return types.ErrFailedToObtainMsgSigners.Wrapf("msg %s, err %v", sdk.MsgTypeURL(msg), err)
		}

		for _, signer := range signers {

			// Make sure that the message signer is equivalent to the mapped Sequencer address of the sender on Ethereum
			if mappedSequencerAddr.String() != sdk.AccAddress(signer).String() {
				return types.ErrInvalidSigner.Wrapf(
					"expected %s, got %s", mappedSequencerAddr.String(), sdk.AccAddress(signer).String(),
				)
			}
		}
	}

	return nil
}

// ExecuteMsg attempts to execute an authorized message originating from Ethereum
func (k Keeper) executeMsg(ctx sdk.Context, msg sdk.Msg) error {
	handler := k.router.Handler(msg)
	if handler == nil {
		return types.ErrInvalidMsgHandlerRoute
	}

	res, err := handler(ctx, msg)
	if err != nil {
		return err
	}

	// The sdk msg handler creates a new EventManager, so events must be correctly propagated back to current context
	ctx.EventManager().EmitEvents(res.GetEvents())

	// Each individual sdk.Result has exactly one Msg response.
	msgResponse := res.MsgResponses[0]
	if msgResponse == nil {
		return types.ErrNilMsgResponse.Wrapf("%s", sdk.MsgTypeURL(msg))
	}

	return nil
}
