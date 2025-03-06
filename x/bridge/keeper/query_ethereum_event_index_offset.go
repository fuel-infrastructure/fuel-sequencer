package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (k Keeper) EthereumEventIndexOffset(goCtx context.Context, req *types.QueryGetEthereumEventIndexOffsetRequest) (*types.QueryGetEthereumEventIndexOffsetResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	val, found := k.GetEthereumEventIndexOffset(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetEthereumEventIndexOffsetResponse{Offset: val}, nil
}
