package keeper

import (
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// ProcessEthereumEvents processes the messages that we have in store
func (k Keeper) ProcessEthereumEvents(ctx sdk.Context) {

	// Firstly retrieve the blockHeight we have last processed.
	lastEthereumBlockSynced := k.MustGetLastEthereumBlockSynced(ctx)

	// Secondly retireve the events and the last block height processed.
	ethEventsTx, ok := k.GetEthEventsTx(ctx, lastEthereumBlockSynced.Uint64())
	if !ok {
		// If no events are found that this block height then we can stop here.
		return
	}

	// Get the params as they are needed for the denom.
	params := k.GetParams(ctx)
	supplyDelta := k.MustGetSupplyDeltaInfo(ctx)

	// Otherwise we being processing these events according to event type.
	for _, event := range ethEventsTx.Events {

		// Unmarshal the parsed event if possible.
		parsedEvent, err := event.UnmarshalParsedEvent()
		if err != nil {
			// If an event cannot be unmarshalled to ParsedEvent ignore it and move on to the next event as there
			// might be something very suspicious with the event.
			k.Logger().Error("Bridge EndBlock: could not unmarshal to ParsedEvent", "event", event.String(), "err", err)
			continue
		}

		switch pe := parsedEvent.(type) {
		case *sidecartypes.SendToSequencerEvent:
			k.processSendToSequencerEvent(ctx, pe, params, &supplyDelta)
		case *sidecartypes.AuthorizeEvent:
			continue
		default:
			continue
		}

	}

	// Set the supply delta info into state.
	k.SetSupplyDeltaInfo(ctx, supplyDelta)
}

// processSendToSequencerEvent processes the send to sequencer events queried from the sidecar.
// Deposit message cannot fail. If a failure can occur, we should consider minting the tokens
// anyways and storing them in the governance address. The only time it can fail is if
// we cannot parse the `Amount` of tokens as we won't know how many tokens have been processed.
func (k Keeper) processSendToSequencerEvent(
	ctx sdk.Context,
	sendEvent *sidecartypes.SendToSequencerEvent,
	params types.Params,
	supplyDeltaInfo *types.SupplyDeltaInfo,
) {

	// TODO: Check that the contract address matches the stored proxy contract address.

	// Validate the send to sequencer event.
	if err := sendEvent.ValidateBasic(); err != nil {
		k.Logger().Error("Bridge EndBlock: could not validate event", "err", err)
	}

	// Parse the data accordingly.
	amount, success := sdkmath.NewIntFromString(sendEvent.Amount)
	if !success {
		k.Logger().Error("Bridge EndBlock: could not unmarshal amount to int from string")
	}
	totalCoins := sdk.NewCoins(sdk.NewCoin(params.BridgeDenom, amount))

	// Parse the vesting duration
	vesting, err := time.ParseDuration(sendEvent.Duration)
	if err != nil {
		k.Logger().Error("Bridge EndBlock: failed to process send to sequencer duration from string", "err", err)
	}

	// If a `To` address was not specified send tokens to the `From` Ethereum Address.
	if len(strings.TrimSpace(sendEvent.To)) == 0 {
		err := k.depositFromEthereum(ctx, sendEvent.From, vesting, totalCoins)
		if err != nil {
			k.Logger().Error("Bridge EndBlock: failed to deposit from ethereum", "err", err)
		}
	} else {
		// Otherwise process the To from a string to an AccAddress type.
		sequencerAddr, err := sdk.AccAddressFromBech32(sendEvent.To)
		if err != nil {
			k.Logger().Error("Bridge EndBlock: to is not a valid Bech32 address", "err", err)
		}

		// TODO Apply logic based on whether the Ethereum sender is the owner of the recipient address

		// Otherwise mint and send the coins to the specified user.
		err = k.bankKeeper.MintCoins(ctx, types.ModuleName, totalCoins)
		if err != nil {
			k.Logger().Error("Bridge EndBlock: failed to mint tokens to module", "err", err)
		}

		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sequencerAddr, totalCoins)
		if err != nil {
			k.Logger().Error("Bridge EndBlock: failed to send tokens from module to account", "err", err)
		}
	}

	// Apply negative offset to supply delta offset
	supplyDeltaInfo.Offset = supplyDeltaInfo.Offset.Sub(amount)

	// TODO do we store the supplyDeltaInfo or do we just keep adding the offset?
	return
}
