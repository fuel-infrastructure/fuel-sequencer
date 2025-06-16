package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

var _ types.QueryServer = Keeper{}

// State implements the Query/State gRPC method
func (k Keeper) State(goCtx context.Context, req *types.QueryStateRequest) (*types.QueryStateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryStateResponse{State: k.GetState(ctx)}, nil
}
