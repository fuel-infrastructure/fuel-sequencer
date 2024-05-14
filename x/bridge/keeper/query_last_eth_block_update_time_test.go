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

func TestLastEthBlockUpdateTimeQuery(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	item := createTestLastEthBlockUpdateTime(keeper, ctx)
	tests := []struct {
		desc     string
		request  *types.QueryGetLastEthBlockUpdateTimeRequest
		response *types.QueryGetLastEthBlockUpdateTimeResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetLastEthBlockUpdateTimeRequest{},
			response: &types.QueryGetLastEthBlockUpdateTimeResponse{LastEthBlockUpdateTime: item},
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.LastEthBlockUpdateTime(ctx, tc.request)
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
