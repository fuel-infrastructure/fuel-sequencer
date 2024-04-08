package testsuite

import (
	"context"
	"fmt"

	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *E2ETestSuite) QueryBridgeParams(ctx context.Context) *bridgetypes.Params {
	queryClient := s.getGRPCClients().BridgeQueryClient
	res, err := queryClient.Params(ctx, &bridgetypes.QueryParamsRequest{})
	s.Require().NoError(err)

	return &res.Params
}

func (s *E2ETestSuite) QueryLastEthereumBlockSynced(ctx context.Context) uint64 {
	queryClient := s.getGRPCClients().BridgeQueryClient
	res, err := queryClient.LastEthereumBlockSynced(ctx, &bridgetypes.QueryGetLastEthereumBlockSyncedRequest{})
	s.Require().NoError(err)

	return res.Block
}

func (s *E2ETestSuite) PollForEthereumEventIndexOffset(
	ctx context.Context, deltaBlocks uint64, ethereumEventIndexOffset uint64,
) {
	h, err := s.Chain.FuelSequencerHeight(ctx)
	s.Require().NoError(err)

	s.T().Log(fmt.Sprintf("Polling for Ethereum event index offset %d", ethereumEventIndexOffset))

	doPoll := func(ctx context.Context, height uint64) (any, error) {
		offset, err := s.Chain.grpcClients.BridgeQueryClient.EthereumEventIndexOffset(ctx,
			&bridgetypes.QueryGetEthereumEventIndexOffsetRequest{},
		)
		if err != nil {
			return nil, err
		}
		if offset.Offset != ethereumEventIndexOffset {
			return nil, fmt.Errorf("offset (%d) does not match expected: (%d)", offset.Offset, ethereumEventIndexOffset)
		}
		return nil, nil
	}

	bp := BlockPoller[any]{CurrentHeight: s.Chain.FuelSequencerHeight, PollFunc: doPoll}
	_, err = bp.DoPoll(ctx, h, h+deltaBlocks)
	s.Require().NoError(err, "exact offset not found in expected number of blocks")
}
