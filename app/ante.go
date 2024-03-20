package app

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
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
		NewInjectedMessagesDecorator(bridgeKeeper),
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

type InjectedMessagesDecorator struct {
	bridgeKeeper bridgekeeper.Keeper
}

func NewInjectedMessagesDecorator(bridgeKeeper bridgekeeper.Keeper) InjectedMessagesDecorator {
	return InjectedMessagesDecorator{
		bridgeKeeper: bridgeKeeper,
	}
}

// AnteHandle implements the AnteHandler decorator for MsgSupplyDelta. If an error is returned from AnteHandle during
// CheckTx, the Tx will get rejected immediately and will not be inserted in the mempool/block.
func (imd InjectedMessagesDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {

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
	if err = msgSupplyDelta.ValidateBasic(); err != nil {
		return ctx, err
	}

	// Get SupplyDeltaPeriod
	supplyDeltaPeriod := imd.bridgeKeeper.GetParams(ctx).SupplyDeltaPeriod
	if supplyDeltaPeriod == 0 {
		return ctx, errors.New("SupplyDeltaPeriod cannot be zero")
	}

	// MsgSupplyDelta can only be injected at specific height intervals. We will reject the Tx if a MsgSupplyDelta is
	// injected at an unexpected height as this must be user-generated
	if uint64(ctx.BlockHeight())%supplyDeltaPeriod != 0 {
		return ctx, fmt.Errorf("MsgSupplyDelta not expected at height %d", ctx.BlockHeight())
	}

	// TODO: Add flag and remove at endblocker

	// Other Ante decorators won't execute if we reach this stage
	return ctx, nil
}
