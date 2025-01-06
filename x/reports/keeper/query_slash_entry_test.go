package keeper_test

import (
	"testing"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/sample"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestQuerySlashEntry(t *testing.T) {
	keeper, ctx := keepertest.ReportsKeeper(t)
	slashReports := keepertest.CreateNSlashReport(keeper, ctx, 2)
	orderedSlashReports := keepertest.OrderSlashReportsLexicographically(slashReports)
	tests := []struct {
		desc     string
		request  *types.QueryGetSlashEntryRequest
		response *types.QueryGetSlashEntryResponse
		err      error
	}{
		{
			desc: "Get first SlashReport's entry 0 from state",
			request: &types.QueryGetSlashEntryRequest{
				Height:           orderedSlashReports[0].Height,
				DelegatorAddress: orderedSlashReports[0].Entries[0].DelegatorAddress,
				ValidatorAddress: orderedSlashReports[0].Entries[0].ValidatorAddress,
			},
			response: &types.QueryGetSlashEntryResponse{SlashEntry: orderedSlashReports[0].Entries[0]},
		},
		{
			desc: "Get second SlashReport's entry 0 from state",
			request: &types.QueryGetSlashEntryRequest{
				Height:           orderedSlashReports[1].Height,
				DelegatorAddress: orderedSlashReports[1].Entries[0].DelegatorAddress,
				ValidatorAddress: orderedSlashReports[1].Entries[0].ValidatorAddress,
			},
			response: &types.QueryGetSlashEntryResponse{SlashEntry: orderedSlashReports[1].Entries[0]},
		},
		{
			desc: "Get second SlashReport's entry 1 from state",
			request: &types.QueryGetSlashEntryRequest{
				Height:           orderedSlashReports[1].Height,
				DelegatorAddress: orderedSlashReports[1].Entries[1].DelegatorAddress,
				ValidatorAddress: orderedSlashReports[1].Entries[1].ValidatorAddress,
			},
			response: &types.QueryGetSlashEntryResponse{SlashEntry: orderedSlashReports[1].Entries[1]},
		},
		{
			desc: "KeyNotFound - wrong height",
			request: &types.QueryGetSlashEntryRequest{
				Height:           100000, // Only two slash reports in state, one with height 1 and the other with height 2
				DelegatorAddress: orderedSlashReports[0].Entries[0].DelegatorAddress,
				ValidatorAddress: orderedSlashReports[0].Entries[0].ValidatorAddress,
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "KeyNotFound - wrong delegator",
			request: &types.QueryGetSlashEntryRequest{
				Height:           orderedSlashReports[0].Height,
				DelegatorAddress: sample.AccAddress(),
				ValidatorAddress: orderedSlashReports[0].Entries[0].ValidatorAddress,
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "KeyNotFound - wrong validator",
			request: &types.QueryGetSlashEntryRequest{
				Height:           orderedSlashReports[1].Height,
				DelegatorAddress: orderedSlashReports[0].Entries[0].DelegatorAddress,
				ValidatorAddress: sample.ValAddress(),
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
			response, err := keeper.SlashEntry(ctx, tc.request)
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
