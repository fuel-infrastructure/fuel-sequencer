package keeper

import (
	"context"
	"fmt"

	"github.com/fuel-infrastructure/fuel-sequencer/utils"
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

	if err := utils.ValidateBech32Address(req.DelegatorAddress); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid delegator address: %s", err.Error()))
	}
	if err := utils.ValidateBech32ValAddress(req.ValidatorAddress); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid validator address: %s", err.Error()))
	}

	val, found := k.GetSlashEntry(ctx, req.Height, req.DelegatorAddress, req.ValidatorAddress)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetSlashEntryResponse{SlashEntry: val}, nil
}
