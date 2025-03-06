package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (k Keeper) LastEthereumNonce(
	goCtx context.Context,
	req *types.QueryGetLastEthereumNonceRequest,
) (*types.QueryGetLastEthereumNonceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	val, found := k.GetLastEthereumNonce(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetLastEthereumNonceResponse{Nonce: val.String()}, nil
}
