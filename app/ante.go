package app

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	bridgekeeper "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
)

func NewAnteHandler(options ante.HandlerOptions, bridgeKeeper bridgekeeper.Keeper) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, sdkerrors.ErrLogic.Wrap("sign mode handler is required for ante builder")
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		NewCustomDecorator(bridgeKeeper),
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

type CustomDecorator struct {
	bridgeKeeper bridgekeeper.Keeper
}

func NewCustomDecorator(bridgeKeeper bridgekeeper.Keeper) CustomDecorator {
	return CustomDecorator{
		bridgeKeeper: bridgeKeeper,
	}
}

// AnteHandle TODO
func (d CustomDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler,
) (sdk.Context, error) {

	// There's no special transactions at genesis, and they are only at FinalizeBlock.
	if ctx.BlockHeight() == 0 || ctx.ExecMode() != sdk.ExecModeFinalize {
		return next(ctx, tx, simulate)
	}

	// Set an infinite gas meter temporarily since we might be processing a special or injected transaction.
	cachedGasMeter := ctx.GasMeter()
	ctx = ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())

	// If the index does not exist, then this must be the first transaction that will set the index.
	// We're done from the AnteHandler and can keep the infinite gas meter for the message handler.
	index, found := d.bridgeKeeper.GetIndex(ctx)
	if !found {
		return ctx, nil
	}

	// If the AnteHandler has not seen all injected transactions, this must be an injected transaction.
	// We're done from the AnteHandler and can keep the infinite gas meter for the respective message handler.
	if index.NumInjectedTxsAnte < index.NumInjectedTxsTotal {

		// The AnteHandler has seen an injected transaction.
		index.NumInjectedTxsAnte += 1
		d.bridgeKeeper.SetIndex(ctx, index)

		return ctx, nil
	}

	// Revert the gas meter
	ctx = ctx.WithGasMeter(cachedGasMeter)

	return next(ctx, tx, simulate)
}
