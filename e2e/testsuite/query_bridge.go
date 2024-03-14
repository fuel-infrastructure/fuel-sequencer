package testsuite

import (
	"context"
	"strconv"

	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *E2ETestSuite) QueryBridgeParams(ctx context.Context) *bridgetypes.Params {
	queryClient := s.getGRPCClients().BridgeQueryClient
	res, err := queryClient.Params(ctx, &bridgetypes.QueryParamsRequest{})
	s.Require().NoError(err)

	return &res.Params
}

func (s *E2ETestSuite) QueryLastEthereumBlockSynced(ctx context.Context) int {
	queryClient := s.getGRPCClients().BridgeQueryClient
	res, err := queryClient.LastEthereumBlockSynced(ctx, &bridgetypes.QueryGetLastEthereumBlockSyncedRequest{})
	s.Require().NoError(err)

	block, err := strconv.Atoi(res.Block)
	s.Require().NoError(err)

	return block
}
