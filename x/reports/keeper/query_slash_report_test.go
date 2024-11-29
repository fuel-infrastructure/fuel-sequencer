package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/nullify"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
)

func TestQuerySlashReport(t *testing.T) {
	keeper, ctx := keepertest.ReportsKeeper(t)
	slashReports := keepertest.CreateNSlashReport(keeper, ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetSlashReportRequest
		response *types.QueryGetSlashReportResponse
		err      error
	}{
		{
			desc: "Get first SlashReport from state",
			request: &types.QueryGetSlashReportRequest{
				Height: slashReports[0].Height,
			},
			response: &types.QueryGetSlashReportResponse{SlashReport: slashReports[0]},
		},
		{
			desc: "Get second SlashReport from state",
			request: &types.QueryGetSlashReportRequest{
				Height: slashReports[1].Height,
			},
			response: &types.QueryGetSlashReportResponse{SlashReport: slashReports[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetSlashReportRequest{
				Height: 100000, // Only two slash reports in state, one with height 1 and the other with height 2
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "request cannot be nil"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.SlashReport(ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response),
					nullify.Fill(response),
				)
			}
		})
	}
}

func TestQuerySlashReportAll(t *testing.T) {
	keeper, ctx := keepertest.ReportsKeeper(t)
	slashReports := keepertest.CreateNSlashReport(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllSlashReportRequest {
		return &types.QueryAllSlashReportRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(slashReports); i += step {
			resp, err := keeper.SlashReportAll(ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.SlashReport), step)
			require.Subset(t,
				nullify.Fill(slashReports),
				nullify.Fill(resp.SlashReport),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(slashReports); i += step {
			resp, err := keeper.SlashReportAll(ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.SlashReport), step)
			require.Subset(t,
				nullify.Fill(slashReports),
				nullify.Fill(resp.SlashReport),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.SlashReportAll(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(slashReports), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(slashReports),
			nullify.Fill(resp.SlashReport),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.SlashReportAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "request cannot be nil"))
	})
}
