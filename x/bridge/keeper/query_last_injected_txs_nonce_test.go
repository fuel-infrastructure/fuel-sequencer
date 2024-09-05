package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/nullify"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func TestQueryLastInjectedTxsNonce(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	value := math.NewInt(10)

	keeper.SetLastInjectedTxsNonce(ctx, value)
	tests := []struct {
		desc     string
		request  *types.QueryGetLastInjectedTxsNonceRequest
		response *types.QueryGetLastInjectedTxsNonceResponse
		err      error
	}{
		{
			desc:     "ValidRequest",
			request:  &types.QueryGetLastInjectedTxsNonceRequest{},
			response: &types.QueryGetLastInjectedTxsNonceResponse{Nonce: value.String()},
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.LastInjectedTxsNonce(ctx, tc.request)
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
