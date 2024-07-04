package keeper

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// processDepositEvent processes the deposit events queried from the sidecar.
// Deposits message cannot fail, so either the handler panics or we store the minted
// tokens in the governance address. The only time processDepositEvent can panic
// is if we cannot parse the `Amount` of tokens as we won't know how many tokens
// have been processed.
func (k Keeper) processDepositEvent(
	ctx sdk.Context,
	depositEvent *types.MsgDepositFromEthereum,
	params *types.Params,
	supplyDeltaInfo *types.SupplyDeltaInfo,
) {
	// Parse the data accordingly.
	amount, success := sdkmath.NewIntFromString(depositEvent.Amount)
	if !success {
		// Panic if we've failed to unmarshal an amount, as something has gone wrong in the entire chain process.
		panic(fmt.Sprintf("could not unmarshal amount to int from string (%s)", depositEvent.Amount))
	}
	tokenToMint := sdk.NewCoin(params.BridgeDenom, amount)
	tokensToMint := sdk.NewCoins(tokenToMint)

	// Check that the Lockup can be converted from a string to sdk.Int
	eventLockup, success := sdkmath.NewIntFromString(depositEvent.Lockup)
	if !success {
		k.Logger().Error("could not unmarshal lockup to int from string", "lockup", depositEvent.Lockup)
		k.mintToGovernanceAddress(ctx, tokenToMint, depositEvent, supplyDeltaInfo)
		return
	}

	// Convert the lockup in seconds to a vesting duration
	vesting := time.Duration(eventLockup.Int64() * 1e9)

	// Check that Depositor is a valid hex address
	if !common.IsHexAddress(depositEvent.Depositor) {
		k.Logger().Error(
			"depositor address is not a valid hex address - minting to governance address instead",
			"event", depositEvent,
		)
		k.mintToGovernanceAddress(ctx, tokenToMint, depositEvent, supplyDeltaInfo)
		return
	}

	// Check that Depositor is not a blocked address
	depositorIsBlocked, err := k.IsAddressBlocked(ctx, depositEvent.Depositor)
	if err != nil {
		k.Logger().Error(
			"failed to check if depositor address is blocked - minting to governance address instead",
			"event", depositEvent,
			"err", err,
		)
		k.mintToGovernanceAddress(ctx, tokenToMint, depositEvent, supplyDeltaInfo)
		return
	} else if depositorIsBlocked {
		k.Logger().Error(
			"depositor address is blocked - minting to governance address instead",
			"event", depositEvent,
		)
		k.mintToGovernanceAddress(ctx, tokenToMint, depositEvent, supplyDeltaInfo)
		return
	}

	// sequencerAddr, i.e. the actual recipient address, is to be determined based on the event data.
	var sequencerAddr sdk.AccAddress

	// Generate a potential sequencer address from the Ethereum 'Depositor' address.
	potentialSequencerAddr, seqErr := k.GenerateSequencerAddressFromEthereumAddress(depositEvent.Depositor)

	// If Recipient is owned by Depositor, send tokens to the address mapped 1-to-1 fom the Depositor Ethereum address.
	if isRecipientOwnedByDepositor(depositEvent.Depositor, depositEvent.Recipient, potentialSequencerAddr.String(), seqErr) {
		sequencerAddr, err = k.generateSequencerAccountFromEthereumDeposit(ctx, depositEvent.Depositor, vesting, tokensToMint)
		if err != nil {
			k.Logger().Error(
				"failed to generate sequencer account from ethereum address - minting to gov address",
				"event", depositEvent, "err", err,
			)
			k.mintToGovernanceAddress(ctx, tokenToMint, depositEvent, supplyDeltaInfo)
			return
		}
	} else {

		// If Recipient is an Ethereum address map it to a Sequencer address, otherwise, generate the sdk.AccAddress from the
		// Bech32 string
		if common.IsHexAddress(depositEvent.Recipient) {
			sequencerAddr, err = k.GenerateSequencerAddressFromEthereumAddress(depositEvent.Recipient)
		} else {
			sequencerAddr, err = sdk.AccAddressFromBech32(depositEvent.Recipient)
		}
		if err != nil {
			k.Logger().Error(
				"recipient is not a valid Bech32 or Hex address - minting to gov address",
				"event", depositEvent, "err", err,
			)
			k.mintToGovernanceAddress(ctx, tokenToMint, depositEvent, supplyDeltaInfo)
			return
		}
	}

	// Otherwise mint and send the coins to the specified user.
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, tokensToMint)
	if err != nil {
		panic(fmt.Sprintf("failed to mint tokens to module; err: %s", err.Error()))
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sequencerAddr, tokensToMint)
	if err != nil {
		k.Logger().Error("failed to send tokens from module to account", "err", err)

		// Do not panic here because a receiver address could be blocked. Instead send the minted tokens to
		// the governance account. If that fails we can then panic because it's a misconfiguration of the modules.
		err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, govtypes.ModuleName, tokensToMint)
		if err != nil {
			panic(fmt.Sprintf("failed to transfer bridge tokens to gov module account err: %s", err.Error()))
		}
	}

	// Apply negative offset to supply delta offset
	supplyDeltaInfo.Offset = supplyDeltaInfo.Offset.Sub(amount)
	k.SetSupplyDeltaInfo(ctx, *supplyDeltaInfo)

	// Emit event once completed
	err = ctx.EventManager().EmitTypedEvent(&types.EventDepositEventProcessed{
		Depositor: depositEvent.Depositor,
		Recipient: sequencerAddr.String(),
		Amount:    tokenToMint,
		Lockup:    vesting.String(),
	})
	if err != nil {
		k.Logger().Error("failed to emit event deposit", "err", err)
	}

	k.Logger().Debug("minted bridge tokens to account", "amount", tokenToMint.Amount, "address", sequencerAddr)
}

// mintToGovernanceAddress mints to the governance address in case of an error in normal processing.
func (k Keeper) mintToGovernanceAddress(
	ctx sdk.Context,
	tokenToMint sdk.Coin,
	depositEvent *types.MsgDepositFromEthereum,
	supplyDeltaInfo *types.SupplyDeltaInfo,
) {

	// tokensToMint is the new coins that will be minted
	tokensToMint := sdk.NewCoins(tokenToMint)

	// Mint bridged tokens to governance module address.
	err := k.bankKeeper.MintCoins(ctx, govtypes.ModuleName, tokensToMint)
	if err != nil {
		panic(fmt.Sprintf("failed to mint bridge tokens to gov module account err: %s", err.Error()))
	}

	// Update supply delta to reflect the minting to the governance address.
	supplyDeltaInfo.Offset = supplyDeltaInfo.Offset.Sub(tokenToMint.Amount)

	// Save supply delta here after processing mints.
	k.SetSupplyDeltaInfo(ctx, *supplyDeltaInfo)

	// Marshal event details to bytes.
	eventDetails, err := depositEvent.Marshal()
	if err != nil {
		panic(fmt.Sprintf("failed to marshal event details into bytes: %s", err.Error()))
	}

	// Emit event once completed.
	err = ctx.EventManager().EmitTypedEvent(&types.EventDepositEventFailed{EventDetails: eventDetails})
	if err != nil {
		k.Logger().Error("failed to emit event deposit", "err", err)
	}

	k.Logger().Warn("minted bridge tokens to governance address", "amount", tokenToMint.Amount)
}
