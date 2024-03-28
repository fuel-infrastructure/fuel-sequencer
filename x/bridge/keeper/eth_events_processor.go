package keeper

import (
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// ProcessEthereumEvents processes the messages that we have in store
func (k Keeper) ProcessEthereumEvents(ctx sdk.Context) {

	// Firstly retrieve the blockHeight we have last processed.
	lastEthereumBlockSynced := k.MustGetLastEthereumBlockSynced(ctx)

	// Secondly retrieve the events and the last block height processed.
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

	k.RemoveEthEventsTx(ctx, lastEthereumBlockSynced.Uint64())
}

// processSendToSequencerEvent processes the send to sequencer events queried from the sidecar.
// Deposit message cannot fail. If a failure can occur, we should consider minting the tokens
// anyway and storing them in the governance address. The only time it can fail is if
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

	// Parse the vesting duration
	vesting, err := time.ParseDuration(sendEvent.Duration)
	if err != nil {
		k.Logger().Error("Bridge EndBlock: failed to process send to sequencer duration from string", "err", err)
		k.mintToGovernanceAddress(ctx, tokenToMint, supplyDeltaInfo)
		return
	}

	// Check that From is a valid hex address
	if !common.IsHexAddress(sendEvent.From) {
		k.Logger().Error("Bridge EndBlock: from is not a valid hex address")
		k.mintToGovernanceAddress(ctx, tokenToMint, supplyDeltaInfo)
		return
	}

	// sequencerAddr to be determined based on data provided in event.
	var sequencerAddr sdk.AccAddress

	// If a `To` address was not specified send tokens to the `From` Ethereum Address.
	if len(strings.TrimSpace(sendEvent.To)) == 0 {
		sequencerAddr, err = k.generateSequencerAccountFromEthereumAddress(ctx, sendEvent.From, vesting, tokensToMint)
		if err != nil {
			k.Logger().Error("Bridge EndBlock: failed to generate sequencer account from ethereum address", "err", err)
			k.mintToGovernanceAddress(ctx, tokenToMint, supplyDeltaInfo)
			return
		}
	} else {

		// Otherwise process the To from a string to an AccAddress type.
		sequencerAddr, err = sdk.AccAddressFromBech32(sendEvent.To)
		if err != nil {
			k.Logger().Error("Bridge EndBlock: to is not a valid Bech32 address", "err", err)
			k.mintToGovernanceAddress(ctx, tokenToMint, supplyDeltaInfo)
			return
		}

		// TODO Apply logic based on whether the Ethereum sender is the owner of the recipient address
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
	supplyDeltaInfo.Offset = supplyDeltaInfo.Offset.Add(amount)

	// We have to save the supply delta here incase we panic at a later deposit.
	k.SetSupplyDeltaInfo(ctx, *supplyDeltaInfo)

	k.Logger().Info("Bridge EndBlock: Minted bridge tokens to account", "amount", tokenToMint.Amount, "address", sequencerAddr)
}

func (k Keeper) mintToGovernanceAddress(ctx sdk.Context, tokenToMint sdk.Coin, supplyDeltaInfo *types.SupplyDeltaInfo) {

	// tokensToMint is the new coins that will be minted
	tokensToMint := sdk.NewCoins(tokenToMint)

	// Mint index tokens to module address.
	err := k.bankKeeper.MintCoins(ctx, types.ModuleName, tokensToMint)
	if err != nil {
		panic(fmt.Errorf("failed to mint bridge tokens to bridge module account err: %s", err))
	}

	// Send the minted tokens from the module address to the governance module address.
	err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, govtypes.ModuleName, tokensToMint)
	if err != nil {
		panic(fmt.Errorf("failed to transfer minted tokens from module address to governance address %s, err: %s", govtypes.ModuleName, err))
	}

	// Update supply delta to reflect the minting to the community pool.
	supplyDeltaInfo.Offset = supplyDeltaInfo.Offset.Add(tokenToMint.Amount)

	// We have to save the supply delta here in case we panic at a later deposit.
	k.SetSupplyDeltaInfo(ctx, *supplyDeltaInfo)

	k.Logger().Info("Minted bridge tokens to community pool", "amount", tokenToMint.Amount)
}
