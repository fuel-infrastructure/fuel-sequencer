package app

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	bridgekeeper "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	sequencingkeeper "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/keeper"
)

func NewAnteHandler(
	options ante.HandlerOptions,
	bridgeKeeper bridgekeeper.Keeper,
	sequencingkeeper sequencingkeeper.Keeper,
) (sdk.AnteHandler, error) {
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
		NewInjectedTxsDecorator(bridgeKeeper),

		// Executed after InjectedTxsDecorator to make sure that we are not applying unnecessary limits to injected txs.
		// Note: Injected txs are not expected to reach this point.
		NewSequencerNativeTxsDecorator(sequencingkeeper),

		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),

		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),

		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}

type InjectedTxsDecorator struct {
	bridgeKeeper bridgekeeper.Keeper
}

func NewInjectedTxsDecorator(bridgeKeeper bridgekeeper.Keeper) InjectedTxsDecorator {
	return InjectedTxsDecorator{
		bridgeKeeper: bridgeKeeper,
	}
}

// AnteHandle implements the AnteHandler decorator for injected transactions.
func (d InjectedTxsDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler,
) (sdk.Context, error) {

	// There's no injected transactions at genesis, and they are only at FinalizeBlock.
	if ctx.BlockHeight() == 0 || ctx.ExecMode() != sdk.ExecModeFinalize {
		return next(ctx, tx, simulate)
	}

	// Set an infinite gas meter from now since we might be processing an injected transaction.
	cachedGasMeter := ctx.GasMeter()
	ctx = ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())

	// If the index does not exist, then this must be the first transaction that will set the index.
	// We're done from the AnteHandler and can keep the infinite gas meter for the message handler.
	//
	// Assumption: first tx always contains a MsgIndex. This should be amended if it is no longer the case.
	index, found := d.bridgeKeeper.GetIndex(ctx)
	if !found {
		return ctx, nil
	}

	// If the AnteHandler has not seen all injected transactions, this must be an injected transaction.
	// Note that the consideration of 'all injected transactions' here covers the MsgSupplyDelta as well.
	//
	// We're done from the AnteHandler and can keep the infinite gas meter for the respective message handler.
	if index.NumInjectedTxsAnte < index.NumInjectedTxsTotal {

		// The AnteHandler has seen an injected transaction.
		index.NumInjectedTxsAnte += 1
		d.bridgeKeeper.SetIndex(ctx, index)

		return ctx, nil
	}

	// Revert the gas meter because if we reach this stage, the tx is not an injected one.
	ctx = ctx.WithGasMeter(cachedGasMeter)

	return next(ctx, tx, simulate)
}

type SequencerNativeTxsDecorator struct {
	sequencingKeeper sequencingkeeper.Keeper
}

func NewSequencerNativeTxsDecorator(sequencingKeeper sequencingkeeper.Keeper) SequencerNativeTxsDecorator {
	return SequencerNativeTxsDecorator{
		sequencingKeeper: sequencingKeeper,
	}
}

// AnteHandle implements a custom AnteHandler decorator for Sequencer-native transactions. This decorator will error
// if the size of the tx exceeds SequencerTxMaxBytes.
// Note: If an error is returned from AnteHandle during CheckTx, the Tx will get rejected immediately and will not be
// inserted in the mempool/block.
func (d SequencerNativeTxsDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler,
) (sdk.Context, error) {
	txSize := uint64(len(ctx.TxBytes()))
	params := d.sequencingKeeper.GetParams(ctx)

	if txSize > params.SequencerTxMaxBytes {
		return ctx, fmt.Errorf("transaction is too large; %d > %d", txSize, params.SequencerTxMaxBytes)
	}

	return next(ctx, tx, simulate)
}
