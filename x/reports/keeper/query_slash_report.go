package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/query"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
)

// initPageRequestDefaults is a copy of the unexported initPageRequestDefaults function from the pagination query types.
// Ref: https://github.com/fuel-infrastructure/cosmos-sdk/blob/v0.50.10/types/query/pagination.go#L140-L160
func initPageRequestDefaults(pageRequest *query.PageRequest) *query.PageRequest {
	// if the PageRequest is nil, use default PageRequest
	if pageRequest == nil {
		pageRequest = &query.PageRequest{}
	}

	pageRequestCopy := *pageRequest
	if len(pageRequestCopy.Key) == 0 {
		pageRequestCopy.Key = nil
	}

	if pageRequestCopy.Limit == 0 {
		pageRequestCopy.Limit = query.DefaultLimit

		// count total results when the limit is zero/not supplied
		pageRequestCopy.CountTotal = true
	}

	return &pageRequestCopy
}

// paginateSlashReports does pagination on all SlashReports in state. We cannot use the SDK's Paginate implementation
// since we store slash entries. The SDK implementation may cause partial slash report data to be retrieved from state.
// Ref to SDK's Paginate: https://github.com/fuel-infrastructure/cosmos-sdk/blob/v0.50.10/types/query/pagination.go#L52
func paginateSlashReports(
	allSlashReports []types.SlashReport,
	pageRequest *query.PageRequest,
	onResult func(report types.SlashReport) error,
) (*query.PageResponse, error) {
	pageRequest = initPageRequestDefaults(pageRequest)

	if pageRequest.Offset > 0 && pageRequest.Key != nil {
		return nil, fmt.Errorf("invalid request, either offset or key is expected, got both")
	}

	var count uint64
	var nextKey []byte

	// Handle reverse if requested
	if pageRequest.Reverse {
		slices.Reverse(allSlashReports)
	}

	if len(pageRequest.Key) != 0 {

		// Find starting position when key is provided
		startHeight := types.ExtractHeightFromSlashEntryKey(pageRequest.Key)
		var startIndex int
		var found bool
		for i, slashReport := range allSlashReports {
			if slashReport.Height == startHeight {
				startIndex = i
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("invalid pagination key: no slash report found for height %d", startHeight)
		}

		for _, slashReport := range allSlashReports[startIndex:] {
			if count == pageRequest.Limit {
				nextKey = types.SlashReportKeyPrefix(slashReport.Height)
				break
			}
			err := onResult(slashReport)
			if err != nil {
				return nil, err
			}

			count++
		}

		return &query.PageResponse{
			NextKey: nextKey,
		}, nil
	}

	end := pageRequest.Offset + pageRequest.Limit

	for _, slashReport := range allSlashReports {
		count++

		if count <= pageRequest.Offset {
			continue
		}
		if count <= end {
			err := onResult(slashReport)
			if err != nil {
				return nil, err
			}
		} else if count == end+1 {
			nextKey = types.SlashReportKeyPrefix(slashReport.Height)

			if !pageRequest.CountTotal {
				break
			}
		}
	}

	res := &query.PageResponse{NextKey: nextKey}
	if pageRequest.CountTotal {
		res.Total = count
	}

	return res, nil
}

func (k Keeper) SlashReportAll(
	ctx context.Context,
	req *types.QueryAllSlashReportRequest,
) (*types.QueryAllSlashReportResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Get all reports first
	allSlashReports := k.GetAllSlashReport(ctx)

	var slashReports []types.SlashReport
	pageRes, err := paginateSlashReports(allSlashReports, req.Pagination, func(report types.SlashReport) error {
		slashReports = append(slashReports, report)
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
