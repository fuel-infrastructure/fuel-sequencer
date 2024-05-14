package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) LastEthBlockUpdateTime(
	goCtx context.Context, req *types.QueryGetLastEthBlockUpdateTimeRequest,
) (*types.QueryGetLastEthBlockUpdateTimeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	val, found := k.GetLastEthBlockUpdateTime(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetLastEthBlockUpdateTimeResponse{LastEthBlockUpdateTime: val}, nil
}
