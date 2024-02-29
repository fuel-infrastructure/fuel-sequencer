package commitments

import (
	"context"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/client"
)

func getNodeStatus(clientCtx client.Context) (*coretypes.ResultStatus, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return &coretypes.ResultStatus{}, err
	}
	return node.Status(context.Background()) // TODO: OK to use context.Background() here?
}

func getBlockHeight(clientCtx client.Context) (int64, error) {
	status, err := getNodeStatus(clientCtx)
	if err != nil {
		return 0, err
	}
	height := status.SyncInfo.LatestBlockHeight
	return height, nil
}

func getBlock(clientCtx client.Context, height *int64) (*coretypes.ResultBlock, error) {
	// get the node
	node, err := clientCtx.GetNode()
	if err != nil {
		return nil, err
	}

	return node.Block(context.Background(), height) // TODO: OK to use context.Background() here?
}

func getBlockResults(clientCtx client.Context, height *int64) (*coretypes.ResultBlockResults, error) {
	// get the node
	node, err := clientCtx.GetNode()
	if err != nil {
		return nil, err
	}

	return node.BlockResults(context.Background(), height) // TODO: OK to use context.Background() here?
}
