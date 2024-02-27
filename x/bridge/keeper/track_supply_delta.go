package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// MintBurnEventProcessor takes an event an extracts a mint and/or burn amount from it.
type MintBurnEventProcessor func(sdk.Context, Keeper, sdk.Event) (mint, burn math.Int)

// UpdateSupplyDeltaFromBeginBlockEvents tries to find mint and burn events that should be tracked. Any observed events
// that are of interest are parsed and applied to the supply delta info.
func (k Keeper) UpdateSupplyDeltaFromBeginBlockEvents(ctx sdk.Context) {

	totalMint := math.ZeroInt()
	totalBurn := math.ZeroInt()

	// Mapping of event type to the processor for it.
	eventProcessors := map[string]MintBurnEventProcessor{
		slashingtypes.EventTypeSlash: getBurnAmountFromSlashEvent,
		minttypes.EventTypeMint:      getMintAmountFromMintEvent,
		// TODO: monitor for gov module burns
		// - https://github.com/cosmos/cosmos-sdk/blob/v0.50.2/x/gov/keeper/tally.go#L18
		// - https://github.com/cosmos/cosmos-sdk/blob/v0.50.2/x/gov/keeper/deposit.go#L184-L186
	}

	// Process events of interest.
	for _, event := range ctx.EventManager().Events() {
		if processEvent, ok := eventProcessors[event.Type]; ok {
			mint, burn := processEvent(ctx, k, event)

			totalMint = totalMint.Add(mint)
			totalBurn = totalBurn.Add(burn)

			ctx.Logger().Debug("processed event", "event_type", event, "mint", mint, "burn", burn)
		}
	}

	// If changes were observed, apply them
	if totalMint.IsPositive() || totalBurn.IsPositive() {

		supplyDeltaInfo, found := k.GetSupplyDeltaInfo(ctx)
		if !found {
			panic("expected to find supply delta info in state")
		}

		supplyDeltaInfo.Mint = supplyDeltaInfo.Mint.Add(totalMint)
		supplyDeltaInfo.Burn = supplyDeltaInfo.Burn.Add(totalBurn)
		k.SetSupplyDeltaInfo(ctx, supplyDeltaInfo)
	}
}

func getMintAmountFromMintEvent(ctx sdk.Context, _ Keeper, event sdk.Event) (mint, burn math.Int) {

	mint = math.ZeroInt()
	burn = math.ZeroInt()

	// Expect mint event.
	if event.Type != minttypes.EventTypeMint {
		panic(fmt.Sprintf("unexpected event type, expected %s got %s", minttypes.EventTypeMint, event.Type))
	}

	// Expect amount attribute.
	attribute, ok := event.GetAttribute(sdk.AttributeKeyAmount)
	if !ok {
		ctx.Logger().Warn("found mint event without amount attribute", "event", event)
		return
	}

	// Expect amount attribute to be parseable into int.
	amountMinted, ok := math.NewIntFromString(attribute.Value)
	if !ok {
		ctx.Logger().Warn("found non-integer amount in mint event", "event", event, "amount", attribute)
		return
	}

	mint = amountMinted
	return
}

func getBurnAmountFromSlashEvent(ctx sdk.Context, k Keeper, event sdk.Event) (mint, burn math.Int) {

	mint = math.ZeroInt()
	burn = math.ZeroInt()

	// Expect slash event.
	if event.Type != slashingtypes.EventTypeSlash {
		panic(fmt.Sprintf("unexpected event type, expected %s got %s", slashingtypes.EventTypeSlash, event.Type))
	}

	// Expect burned coins attribute.
	attribute, ok := event.GetAttribute(slashingtypes.AttributeKeyBurnedCoins)
	if !ok {
		ctx.Logger().Warn("found slash event without burned coins attribute", "event", event)
		return
	}

	// Expect burned coins to be parseable into coins.
	amountsBurned, err := sdk.ParseCoinsNormalized(attribute.Value)
	if err != nil {
		ctx.Logger().Warn("found non-coins burned coins in burn event", "event", event, "amount", attribute)
		return
	}

	// Expect the FUEL token.
	amountBurned := amountsBurned.AmountOf("ufuel") // TODO: make this dynamic
	if amountBurned.IsZero() {
		ctx.Logger().Warn("did not find slash denom in burned coins in burn event", "event", event, "amount", attribute)
		return
	}

	burn = amountBurned
	return
}
