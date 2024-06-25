package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (k msgServer) Index(goCtx context.Context, msg *types.MsgIndex) (resp *types.MsgIndexResponse, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err = k.TryExecSpecialMessage(ctx, msg.Authority, func(ctx sdk.Context) error {
		var innerErr error
		if resp, innerErr = k.index(ctx, msg); innerErr != nil {
			resp = &types.MsgIndexResponse{} // normalise
			return innerErr
		}
		return nil
	})

	return resp, err
}

func (k msgServer) index(ctx sdk.Context, msg *types.MsgIndex) (*types.MsgIndexResponse, error) {

	lastBlockSynced := k.MustGetLastEthereumBlockSynced(ctx)
	eventIndexOffset := k.MustGetEthereumEventIndexOffset(ctx)
	params := k.GetParams(ctx)

	// If any problem is found in the MsgIndex, this is an indication of a serious bug.
	err := msg.ValidateBeforeProcessing(lastBlockSynced, eventIndexOffset)
	if err != nil {
		return nil, err
	}

	// If it's time for a MsgSupplyDelta, consider this in the total number of injected txs.
	supplyDeltaPeriod := params.SupplyDeltaPeriod
	if supplyDeltaPeriod == 0 {
		return nil, fmt.Errorf("SupplyDeltaPeriod cannot be zero")
	}
	supplyDeltaCount := uint64(0)
	if uint64(ctx.BlockHeight())%supplyDeltaPeriod == 0 {
		supplyDeltaCount += 1
	}

	// Ensure index was not already set
	if _, found := k.GetIndex(ctx); found {
		return nil, fmt.Errorf("expected to not find Index")
	}

	// It is very important to set the index, so the AnteHandler knows that we've processed the MsgIndex.
	k.SetIndex(ctx, types.Index{
		NumInjectedTxsTotal: msg.NumInjectedEventTxs + supplyDeltaCount,
		NumInjectedTxsAnte:  0, // MsgIndex tx is never factored in because it is just a metadata transaction
		NumFailedSpecialTxs: 0, // No special txs have failed yet
	})

	if msg.NewEthereumBlock {
		k.SetLastEthereumBlockSynced(ctx, msg.BlockNumber)
		k.ResetEthereumEventIndexOffset(ctx)
		k.SetLastEthBlockUpdateTime(ctx, ctx.BlockTime())
	}

	// If no new Ethereum block, but we still received some events, then the block was partially consumed.
	// LastEthBlockUpdateTime is also updated to prevent sync issues with partially synced Ethereum blocks
	// when BlockTime exceeds MaxEthBlockUpdateDelay.
	if !msg.NewEthereumBlock && msg.NumInjectedEventTxs > 0 {
		newOffset := eventIndexOffset + msg.NumInjectedEventTxs
		k.SetEthereumEventIndexOffset(ctx, newOffset)
		k.SetLastEthBlockUpdateTime(ctx, ctx.BlockTime())
	}

	return &types.MsgIndexResponse{}, nil
}
