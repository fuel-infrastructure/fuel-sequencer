package keeper

import (
	"context"

	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) SlashEntry(
	ctx context.Context,
	req *types.QueryGetSlashEntryRequest,
) (*types.QueryGetSlashEntryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	val, found := k.GetSlashEntry(ctx, req.Height, req.DelegatorAddress, req.ValidatorAddress)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetSlashEntryResponse{SlashEntry: val}, nil
}
