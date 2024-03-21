package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/nullify"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func TestSupplyDeltaProcessedQuery(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	item := createTestSupplyDeltaProcessed(keeper, ctx)
	tests := []struct {
		desc     string
		request  *types.QueryGetSupplyDeltaProcessedRequest
		response *types.QueryGetSupplyDeltaProcessedResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetSupplyDeltaProcessedRequest{},
			response: &types.QueryGetSupplyDeltaProcessedResponse{SupplyDeltaProcessed: item},
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.SupplyDeltaProcessed(ctx, tc.request)
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
