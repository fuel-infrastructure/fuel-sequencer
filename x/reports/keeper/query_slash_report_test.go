package keeper_test

import (
	"slices"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
)

func TestQuerySlashReport(t *testing.T) {
	keeper, ctx := keepertest.ReportsKeeper(t)
	slashReports := keepertest.CreateNSlashReport(keeper, ctx, 2)
	orderedSlashReports := keepertest.OrderSlashReportsLexicographically(slashReports)
	tests := []struct {
		desc     string
		request  *types.QueryGetSlashReportRequest
		response *types.QueryGetSlashReportResponse
		err      error
	}{
		{
			desc: "Get first SlashReport from state",
			request: &types.QueryGetSlashReportRequest{
				Height: orderedSlashReports[0].Height,
			},
			response: &types.QueryGetSlashReportResponse{SlashReport: orderedSlashReports[0]},
		},
		{
			desc: "Get second SlashReport from state",
			request: &types.QueryGetSlashReportRequest{
				Height: orderedSlashReports[1].Height,
			},
			response: &types.QueryGetSlashReportResponse{SlashReport: orderedSlashReports[1]},
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
					tc.response,
					response,
				)
			}
		})
	}
}

func TestQuerySlashReportAll(t *testing.T) {
	keeper, ctx := keepertest.ReportsKeeper(t)
	slashReports := keepertest.CreateNSlashReport(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool, reverse bool) *types.QueryAllSlashReportRequest {
		return &types.QueryAllSlashReportRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
				Reverse:    reverse,
			},
		}
	}

	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(slashReports); i += step {
			resp, err := keeper.SlashReportAll(ctx, request(nil, uint64(i), uint64(step), false, false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.SlashReport), step)
			require.Subset(t,
				keepertest.OrderSlashReportsLexicographically(slashReports),
				resp.SlashReport,
			)
		}
	})

	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(slashReports); i += step {
			resp, err := keeper.SlashReportAll(ctx, request(next, 0, uint64(step), false, false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.SlashReport), step)
			require.Subset(t,
				keepertest.OrderSlashReportsLexicographically(slashReports),
				resp.SlashReport,
			)
			next = resp.Pagination.NextKey
		}
	})

	t.Run("Reverse", func(t *testing.T) {
		step := 2
		var next []byte
		reversed := make([]types.SlashReport, len(slashReports))
		copy(reversed, keepertest.OrderSlashReportsLexicographically(slashReports))
		slices.Reverse(reversed)

		for i := 0; i < len(slashReports); i += step {
			resp, err := keeper.SlashReportAll(ctx, request(next, 0, uint64(step), false, true))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.SlashReport), step)
			require.Subset(t,
				reversed,
				resp.SlashReport,
			)
			next = resp.Pagination.NextKey
		}
	})

	t.Run("StartFromKey", func(t *testing.T) {
		startKey := types.SlashReportKeyPrefix(slashReports[2].Height)
		resp, err := keeper.SlashReportAll(ctx, request(startKey, 0, 2, false, false))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.SlashReport))
		require.Equal(t, slashReports[2].Height, resp.SlashReport[0].Height)
	})

	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.SlashReportAll(ctx, request(nil, 0, 0, true, false))
		require.NoError(t, err)
		require.Equal(t, len(slashReports), int(resp.Pagination.Total))
		require.Equal(t,
			keepertest.OrderSlashReportsLexicographically(slashReports),
			resp.SlashReport,
		)
	})

	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.SlashReportAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "request cannot be nil"))
	})

	t.Run("InvalidOffsetAndKey", func(t *testing.T) {
		_, err := keeper.SlashReportAll(ctx, request([]byte("key"), 1, 1, false, false))
		require.Error(t, err)
	})

	t.Run("EmptyStateWithOffset", func(t *testing.T) {
		// Clear the state first by defining a new keeper
		keeper, ctx = keepertest.ReportsKeeper(t)

		resp, err := keeper.SlashReportAll(ctx, request(nil, 0, 2, true, false))
		require.NoError(t, err)
		require.Empty(t, resp.SlashReport)
		require.Equal(t, uint64(0), resp.Pagination.Total)
		require.Nil(t, resp.Pagination.NextKey)
	})

	t.Run("EmptyStateWithKey", func(t *testing.T) {
		// Clear the state first by defining a new keeper
		keeper, ctx = keepertest.ReportsKeeper(t)

		// Try to use a key in empty state
		startKey := types.SlashReportKeyPrefix(uint64(123))
		_, err := keeper.SlashReportAll(ctx, request(startKey, 0, 2, false, false))
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid pagination key")
	})
}
