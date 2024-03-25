package keeper

import (
	"fmt"
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
			err = k.processSendToSequencerEvent(ctx, pe, params)
		case *sidecartypes.AuthorizeEvent:
			continue
		default:
			continue
		}

		fmt.Println(err)
	}
}

// processSendToSequencerEvent processes the send to sequencer events queried from the sidecar.
// Deposit message cannot fail. If a failure can occur, we should consider minting the tokens
// anyways and storing them in the governance address.
func (k Keeper) processSendToSequencerEvent(
	ctx sdk.Context,
	sendEvent *sidecartypes.SendToSequencerEvent,
	params types.Params,
) error {

	// TODO: Check that the contract address matches the stored proxy contract address.

	// Validate the send to sequencer event.
	if err := sendEvent.ValidateBasic(); err != nil {
		return err
	}

	// Parse the data accordingly.
	amount, success := sdkmath.NewIntFromString(sendEvent.Amount)
	if !success {
		return fmt.Errorf("failed to process send to sequencer amount to int from string")
	}
	totalCoins := sdk.NewCoins(sdk.NewCoin(params.BridgeDenom, amount))

	// Parse the vesting duration
	vesting, err := time.ParseDuration(sendEvent.Duration)
	if err != nil {
		return fmt.Errorf("failed to process send to sequencer duration from string: %s", err)
	}

	// If a `To` address was not specified send tokens to the `From` Ethereum Address.
	if len(strings.TrimSpace(sendEvent.To)) == 0 {
		err := k.depositFromEthereum(ctx, sendEvent.From, vesting, totalCoins)
		if err != nil {
			return err
		}

		return nil
	}

	// Otherwise process the To from a string to an AccAddress type.
	sequencerAddr, err := sdk.AccAddressFromBech32(sendEvent.To)
	if err != nil {
		return fmt.Errorf("to is not a valid Bech32 address: %w", err)
	}

	// TODO Apply logic based on whether the Ethereum sender is the owner of the recipient address

	// Otherwise mint and send the coins to the specified user.
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, totalCoins)
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sequencerAddr, totalCoins)
	if err != nil {
		return err
	}

	return nil
}
