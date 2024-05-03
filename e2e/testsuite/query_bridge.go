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
	ctx context.Context, deltaBlocks uint64, offset uint64,
) {
	h, err := s.Chain.FuelSequencerHeight(ctx)
	s.Require().NoError(err)

	s.T().Log(fmt.Sprintf("Polling for Ethereum event index offset %d", offset))

	doPoll := func(ctx context.Context, height uint64) (any, error) {
		resp, err := s.Chain.grpcClients.BridgeQueryClient.EthereumEventIndexOffset(ctx,
			&bridgetypes.QueryGetEthereumEventIndexOffsetRequest{},
		)
		if err != nil {
			return nil, err
		}
		if resp.Offset != offset {
			return nil, fmt.Errorf("offset (%d) does not match expected: (%d)", resp.Offset, offset)
		}
		return nil, nil
	}

	bp := BlockPoller[any]{CurrentHeight: s.Chain.FuelSequencerHeight, PollFunc: doPoll}
	_, err = bp.DoPoll(ctx, h, h+deltaBlocks)
	s.Require().NoError(err, fmt.Errorf("exact offset (%d) not found in expected number of blocks", offset))
}

func (s *E2ETestSuite) PollForLastEthereumBlockSynced(
	ctx context.Context, deltaBlocks uint64, block uint64,
) {
	h, err := s.Chain.FuelSequencerHeight(ctx)
	s.Require().NoError(err)

	s.T().Log(fmt.Sprintf("Polling for last Ethereum block synced %d", block))

	doPoll := func(ctx context.Context, height uint64) (any, error) {
		resp, err := s.Chain.grpcClients.BridgeQueryClient.LastEthereumBlockSynced(ctx,
			&bridgetypes.QueryGetLastEthereumBlockSyncedRequest{},
		)
		if err != nil {
			return nil, err
		}
		if resp.Block != block {
			return nil, fmt.Errorf("last block synced (%d) does not match expected: (%d)", resp.Block, block)
		}
		return nil, nil
	}

	bp := BlockPoller[any]{CurrentHeight: s.Chain.FuelSequencerHeight, PollFunc: doPoll}
	_, err = bp.DoPoll(ctx, h, h+deltaBlocks)
	s.Require().NoError(err, fmt.Errorf("exact last block synced %d not found in expected number of blocks", block))
}
