package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (k msgServer) SupplyDelta(goCtx context.Context, msg *types.MsgSupplyDelta) (*types.MsgSupplyDeltaResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Confirm that the msg signer is the bridge module's authority address (governance).
	if k.GetAuthority() != msg.Authority {
		return nil, types.ErrInvalidSigner.Wrapf(
			"invalid authority; expected %s, got %s", k.GetAuthority(), msg.Authority,
		)
	}

	// Confirm that BridgeParams.SupplyDeltaPeriod is non-zero, otherwise we can't calculate the expected height at
	// which a MsgSupplyDelta is to be sent.
	supplyDeltaPeriod := k.GetParams(ctx).SupplyDeltaPeriod
	if supplyDeltaPeriod == 0 {
		return nil, types.ErrInvalidSupplyDeltaPeriod.Wrapf("SupplyDeltaPeriod cannot be zero")
	}

	// Confirm that MsgSupplyDelta was injected at the correct height.
	blockHeight := ctx.BlockHeight()
	if (uint64(blockHeight) % supplyDeltaPeriod) != 0 {
		return nil, types.ErrUnexpectedOperation.Wrapf(
			"MsgSupplyDelta not expected at height %d", blockHeight,
		)
	}

	// Increment LastEthereumNonce and get the result so that it is added to MsgSupplyDeltaResponse.
	nonce := k.MustGetNextEthereumNonce(ctx)

	// Calculate the supply delta to be reported.
	supplyDeltaInfo := k.MustGetSupplyDeltaInfo(ctx)
	supplyDelta := supplyDeltaInfo.Delta.Add(supplyDeltaInfo.Offset)

	// If the supply delta is zero, then there is either nothing to report to Ethereum or MsgSupplyDelta was submitted
	// at a valid height by a user. We should fail in both scenarios.
	if supplyDelta.IsZero() {
		return nil, types.ErrInvalidSupplyDeltaValue.Wrapf("cannot report 0 supply delta to Ethereum")
	}

	// Reset SupplyDeltaInfo
	k.MustResetSupplyDeltaInfo(ctx)

	// Emit event
	err := ctx.EventManager().EmitTypedEvent(&types.EventSupplyDeltaReported{SupplyDelta: supplyDelta, Nonce: nonce})
	if err != nil {
		return nil, err
	}

	return &types.MsgSupplyDeltaResponse{
		Nonce:       nonce,
		SupplyDelta: supplyDelta,
	}, nil
}
