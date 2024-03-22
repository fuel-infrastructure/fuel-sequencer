package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/nullify"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func TestQueryEthEventsTxByBlockNumber(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	events := []*sidecartypes.Event{
		{
			EventType: "authorize",
			Data:      []byte("auth"),
		},
	}
	ethEventsTx := types.EthEventsTx{
		Events:           events,
		AdvanceSequencer: true,
		NewEthereumBlock: true,
		BlockNumber:      math.NewInt(10),
	}

	keeper.SetEthEventsTx(ctx, ethEventsTx)

	tests := []struct {
		desc     string
		request  *types.QueryGetEthEventsTxByBlockNumberRequest
		response *types.QueryGetEthEventsTxByBlockNumberResponse
		err      error
	}{
		{
			desc: "ValidRequest",
			request: &types.QueryGetEthEventsTxByBlockNumberRequest{
				BlockNumber: "10",
			},
			response: &types.QueryGetEthEventsTxByBlockNumberResponse{
				EthEventsTx: &ethEventsTx,
			},
		},
		{
			desc: "InvalidRequest",
			request: &types.QueryGetEthEventsTxByBlockNumberRequest{
				BlockNumber: "abc",
			},
			err: status.Error(codes.InvalidArgument, "could not convert block number to uint64: strconv.ParseUint: parsing \"abc\": invalid syntax"),
		},
		{
			desc: "BlockNotFound",
			request: &types.QueryGetEthEventsTxByBlockNumberRequest{
				BlockNumber: "999",
			},
			err: status.Error(codes.NotFound, "EthEventsTx not found for block number"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.EthEventsTxByBlockNumber(ctx, tc.request)
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
