package service

import (
	"context"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
)

// The code in this file is similar to cosmos-sdk@v0.50.6/client/grpc/cmtservice/block.go

func getLatestBlockHeight(ctx context.Context, clientCtx client.Context) (int64, error) {
	status, err := cmtservice.GetNodeStatus(ctx, clientCtx)
	if err != nil {
		return 0, err
	}
	height := status.SyncInfo.LatestBlockHeight
	return height, nil
}

func getCommit(ctx context.Context, clientCtx client.Context, height *int64) (*coretypes.ResultCommit, error) {
	// get the node
	node, err := clientCtx.GetNode()
	if err != nil {
		return nil, err
	}

	return node.Commit(ctx, height)
}

func getBlockResults(
	ctx context.Context, clientCtx client.Context, height *int64,
) (*coretypes.ResultBlockResults, error) {
	// get the node
	node, err := clientCtx.GetNode()
	if err != nil {
		return nil, err
	}

	return node.BlockResults(ctx, height)
}
