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

	// MsgSupplyDelta Txs will contain only one message.
	msgs := tx.GetMsgs()
	if len(msgs) != 1 {
		return next(ctx, tx, simulate)
	}

	// If the message is not a MsgSupplyDelta continue with the other Ante decorators.
	msg := msgs[0]
	if sdk.MsgTypeURL(msg) != sdk.MsgTypeURL(&bridgetypes.MsgSupplyDelta{}) {
		return next(ctx, tx, simulate)
	}

	// If the message cannot be parsed into MsgSupplyDelta continue with the other Ante decorators.
	msgSupplyDelta, ok := msg.(*bridgetypes.MsgSupplyDelta)
	if !ok {
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
	d.bridgeKeeper.SetSupplyDeltaProcessed(ctx, bridgetypes.SupplyDeltaProcessed{Processed: true})

	// Reset the gas meter to its original state. We can't use a defer function because this doesn't work well with
	// decorators.
	ctx = ctx.WithGasMeter(cachedGasMeter)

	// Other Ante decorators won't execute if we reach this stage
	return ctx, nil
}
