package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (k msgServer) Index(goCtx context.Context, msg *types.MsgIndex) (*types.MsgIndexResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lastBlockSynced := k.MustGetLastEthereumBlockSynced(ctx)
	eventIndexOffset := k.MustGetEthereumEventIndexOffset(ctx)
	params := k.GetParams(ctx)

	// Perform some checks on the injected Ethereum events transaction.
	// If any problem is found, this is an indication of a serious bug.
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

	// It is very important to set the index, so the AnteHandler knows that we've processed the MsgIndex.
	k.SetIndex(ctx, types.Index{
		NumInjectedTxsTotal: msg.NumInjectedTxs + supplyDeltaCount,
		NumInjectedTxsAnte:  0,
	})

	if msg.NewEthereumBlock {
		k.SetLastEthereumBlockSynced(ctx, msg.BlockNumber)
		k.ResetEthereumEventIndexOffset(ctx)
		k.SetLastEthBlockUpdateTime(ctx, ctx.BlockTime())
	}

	// If no new Ethereum block, but we still received some events, then the block was partially consumed.
	if !msg.NewEthereumBlock && msg.NumInjectedTxs > 0 {
		newOffset := eventIndexOffset + msg.NumInjectedTxs
		k.SetEthereumEventIndexOffset(ctx, newOffset)
	}

	return &types.MsgIndexResponse{}, nil
}
