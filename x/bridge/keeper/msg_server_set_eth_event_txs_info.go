package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (k msgServer) SetEthEventTxsInfo(goCtx context.Context, msg *types.EthEventsTx) (*types.MsgSetEthEventTxsInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lastBlockSynced := k.MustGetLastEthereumBlockSynced(ctx)
	eventIndexOffset := k.MustGetEthereumEventIndexOffset(ctx)

	// Perform some checks on the injected Ethereum events transaction.
	// If any problem is found, this is an indication of a serious bug.
	err := msg.ValidateBeforeProcessing(lastBlockSynced, eventIndexOffset)
	if err != nil {
		return nil, fmt.Errorf("eth events tx validation failed: %w", err)
	}

	// It is very important to set the index, so the AnteHandler knows that we've processed the EthEventsTx.
	k.SetEthEventsTxIndex(ctx, types.EthEventsTxIndex{
		NumUnhandledEventTxs: msg.NumInjectedEvents,
	})

	if msg.NewEthereumBlock {
		k.SetLastEthereumBlockSynced(ctx, msg.BlockNumber)
		k.ResetEthereumEventIndexOffset(ctx)
		k.SetLastEthBlockUpdateTime(ctx, ctx.BlockTime())
	}

	// If no new Ethereum block, but we still received some events, then the block was partially consumed.
	if !msg.NewEthereumBlock && msg.NumInjectedEvents > 0 {
		newOffset := eventIndexOffset + msg.NumInjectedEvents
		k.SetEthereumEventIndexOffset(ctx, newOffset)
	}

	return &types.MsgSetEthEventTxsInfoResponse{}, nil
}
