package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) SlashReportAll(
	ctx context.Context,
	req *types.QueryAllSlashReportRequest,
) (*types.QueryAllSlashReportResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	var slashReports []types.SlashReport

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	slashReportStore := prefix.NewStore(store, types.KeyPrefix(types.SlashReportKey))

	pageRes, err := query.Paginate(slashReportStore, req.Pagination, func(key []byte, value []byte) error {
		var slashReport types.SlashReport
		if err := k.cdc.Unmarshal(value, &slashReport); err != nil {
			return err
		}

		slashReports = append(slashReports, slashReport)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllSlashReportResponse{SlashReport: slashReports, Pagination: pageRes}, nil
}

func (k Keeper) SlashReport(
	ctx context.Context,
	req *types.QueryGetSlashReportRequest,
) (*types.QueryGetSlashReportResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	val, found := k.GetSlashReport(ctx, req.Height)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetSlashReportResponse{SlashReport: val}, nil
}
