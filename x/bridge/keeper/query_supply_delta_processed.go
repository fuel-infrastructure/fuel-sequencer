package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) SupplyDeltaProcessed(goCtx context.Context, req *types.QueryGetSupplyDeltaProcessedRequest) (*types.QueryGetSupplyDeltaProcessedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	val, found := k.GetSupplyDeltaProcessed(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetSupplyDeltaProcessedResponse{SupplyDeltaProcessed: val}, nil
}
