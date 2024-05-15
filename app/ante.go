package app

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	bridgekeeper "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func NewAnteHandler(options ante.HandlerOptions, bridgeKeeper bridgekeeper.Keeper) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		NewMsgSetEthEventTxsInfoDecorator(bridgeKeeper),
		NewInjectedEventTxsDecorator(bridgeKeeper),
		NewMsgSupplyDeltaDecorator(bridgeKeeper),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
		ante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}

type MsgSetEthEventTxsInfoDecorator struct {
	bridgeKeeper bridgekeeper.Keeper
}

func NewMsgSetEthEventTxsInfoDecorator(bridgeKeeper bridgekeeper.Keeper) MsgSetEthEventTxsInfoDecorator {
	return MsgSetEthEventTxsInfoDecorator{
		bridgeKeeper: bridgeKeeper,
	}
}

// AnteHandle implements the AnteHandler decorator for MsgSetEthEventTxsInfo. If an error is returned from AnteHandle
// during CheckTx, the Tx will get rejected immediately and will not be inserted in the mempool/block.
func (d MsgSetEthEventTxsInfoDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler,
) (sdk.Context, error) {

	// There's no injected transactions at genesis, and they are only at FinalizeBlock.
	if ctx.BlockHeight() == 0 || ctx.ExecMode() != sdk.ExecModeFinalize {
		return next(ctx, tx, simulate)
	}

	// Override the gas meter with an infinite one to make sure that the message handlers do not run out of gas when
	// they are processing the EthEventsTx. This is safe because we know that it was injected by the consensus logic.
	// We cache the existing one in case we need to revert back to it.
	cachedGasMeter := ctx.GasMeter()
	ctx = ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())

	// Check that we haven't set EthEventsTxIndex yet.
	_, found := d.bridgeKeeper.GetEthEventsTxIndex(ctx)
	if found {
		ctx = ctx.WithGasMeter(cachedGasMeter) // revert
		return next(ctx, tx, simulate)
	}

	// Note: beyond this point, we strictly expect EthEventsTx, and so we should error if anything goes wrong.

	var ethEventsTx bridgetypes.EthEventsTx
	err := ethEventsTx.FromSdkTx(tx)
	if err != nil {
		return ctx, fmt.Errorf("could not get EthEventsTx from tx: %w", err)
	}

	// Confirm that EthEventsTx passes the necessary verification checks and error if not
	if err := ethEventsTx.ValidateBasic(); err != nil {
		return ctx, err
	}

	// Other Ante decorators won't execute if we reach this stage
	return ctx, nil
}

type InjectedEventTxsDecorator struct {
	bridgeKeeper bridgekeeper.Keeper
}

func NewInjectedEventTxsDecorator(bridgeKeeper bridgekeeper.Keeper) InjectedEventTxsDecorator {
	return InjectedEventTxsDecorator{
		bridgeKeeper: bridgeKeeper,
	}
}

// AnteHandle implements the AnteHandler decorator for Ethereum event transactions. If an error is returned from
// AnteHandle during CheckTx, the Tx will get rejected immediately and will not be inserted in the mempool/block.
func (d InjectedEventTxsDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler,
) (sdk.Context, error) {

	// There's no injected transactions at genesis, and they are only at FinalizeBlock.
	if ctx.BlockHeight() == 0 || ctx.ExecMode() != sdk.ExecModeFinalize {
		return next(ctx, tx, simulate)
	}

	// Override the gas meter with an infinite one to make sure that the message handlers do not run out of gas when
	// they are processing the Ethereum event transactions. This is safe because we know that we are processing an event
	// transaction injected by the consensus logic.
	// We cache the existing one in case we need to revert back to it.
	cachedGasMeter := ctx.GasMeter()
	ctx = ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())

	// If we're done processing Ethereum event transactions, proceed to the next decorators.
	// Otherwise, we know for sure that we are processing an Ethereum event transaction.
	eventsIndex := d.bridgeKeeper.MustGetEthEventsTxIndex(ctx)
	if eventsIndex.NumUnhandledEventTxs == 0 {
		ctx = ctx.WithGasMeter(cachedGasMeter) // revert
		return next(ctx, tx, simulate)
	}

	// Note: beyond this point, we strictly expect injected transactions, and so we should error if anything goes wrong.

	// Update events index
	eventsIndex.NumUnhandledEventTxs -= 1
	d.bridgeKeeper.SetEthEventsTxIndex(ctx, eventsIndex)

	// Other Ante decorators won't execute if we reach this stage
	return ctx, nil
}

type MsgSupplyDeltaDecorator struct {
	bridgeKeeper bridgekeeper.Keeper
}

func NewMsgSupplyDeltaDecorator(bridgeKeeper bridgekeeper.Keeper) MsgSupplyDeltaDecorator {
	return MsgSupplyDeltaDecorator{
		bridgeKeeper: bridgeKeeper,
	}
}

// AnteHandle implements the AnteHandler decorator for MsgSupplyDelta. If an error is returned from AnteHandle during
// CheckTx, the Tx will get rejected immediately and will not be inserted in the mempool/block.
func (d MsgSupplyDeltaDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler,
) (sdk.Context, error) {

	// There's no injected transactions at genesis, and they are only at FinalizeBlock.
	if ctx.BlockHeight() == 0 || ctx.ExecMode() != sdk.ExecModeFinalize {
		return next(ctx, tx, simulate)
	}

	var msgSupplyDelta bridgetypes.MsgSupplyDelta
	err := msgSupplyDelta.FromSdkTx(tx)
	if err != nil {
		return next(ctx, tx, simulate)
	}

	// Confirm that msgSupplyDelta passes the necessary verification checks and error if not
	if err := msgSupplyDelta.ValidateBasic(); err != nil {
		return ctx, err
	}

	// Override the gas meter with an infinite one to make sure that the AnteHandler does not run out of gas when it's
	// processing a MsgSupplyDelta. This is safe because any user-initiated MsgSupplyDelta will never be included in a
	// block as below we are erroring when we detect such messages.
	cachedGasMeter := ctx.GasMeter()
	ctx = ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())

	// Get SupplyDeltaPeriod
	supplyDeltaPeriod := d.bridgeKeeper.GetParams(ctx).SupplyDeltaPeriod
	if supplyDeltaPeriod == 0 {
		return ctx, errors.New("SupplyDeltaPeriod cannot be zero")
	}

	// MsgSupplyDelta can only be injected at specific height intervals. We will reject the Tx if a MsgSupplyDelta is
	// injected at an unexpected height as this must be user-generated
	if uint64(ctx.BlockHeight())%supplyDeltaPeriod != 0 {
		return ctx, fmt.Errorf("MsgSupplyDelta not expected at height %d", ctx.BlockHeight())
	}

	// MsgSupplyDelta will be rejected if we have already processed a MsgSupplyDeltaTx. Here we are assuming that
	// module initiated MsgSupplyDeltaTxs are always first of their kind in the block proposal.
	if d.bridgeKeeper.MustGetSupplyDeltaProcessed(ctx).Processed {
		return ctx, errors.New("MsgSupplyDelta already processed in block proposal")
	}

	// Reset the gas meter to its original state. We can't use defer because this doesn't work well with decorators.
	ctx = ctx.WithGasMeter(cachedGasMeter)

	// Other Ante decorators won't execute if we reach this stage
	return ctx, nil
}
