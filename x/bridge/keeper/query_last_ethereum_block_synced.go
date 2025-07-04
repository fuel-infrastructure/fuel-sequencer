package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (k Keeper) LastEthereumBlockSynced(
	goCtx context.Context,
	req *types.QueryGetLastEthereumBlockSyncedRequest,
) (*types.QueryGetLastEthereumBlockSyncedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	val, found := k.GetLastEthereumBlockSynced(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetLastEthereumBlockSyncedResponse{Block: val}, nil
}
