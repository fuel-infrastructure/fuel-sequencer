package service

import (
	"context"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
)

// The code in this file is similar to cosmos-sdk@v0.50.6/client/grpc/cmtservice/block.go

func getLatestBlockHeight(ctx context.Context, node client.CometRPC) (int64, error) {
	status, err := cmtservice.GetNodeStatus(ctx, node)
	if err != nil {
		return 0, err
	}
	height := status.SyncInfo.LatestBlockHeight
	return height, nil
}

func getCommit(ctx context.Context, node client.CometRPC, height *int64) (*coretypes.ResultCommit, error) {
	return node.Commit(ctx, height)
}

func getBlockResults(
	ctx context.Context, node client.CometRPC, height *int64,
) (*coretypes.ResultBlockResults, error) {
	return node.BlockResults(ctx, height)
}
