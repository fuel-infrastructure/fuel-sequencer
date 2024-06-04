package testsuite

import (
	"context"
	"sort"

	"github.com/cometbft/cometbft/libs/bytes"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
)

func (s *E2ETestSuite) GetFuelSequencerHeight(ctx context.Context) (uint64, error) {
	return s.Chain.FuelSequencerHeight(ctx)
}

// Deprecated: Use QueryBridgeCommitment instead.
func (s *E2ETestSuite) GetBridgeCommitment(ctx context.Context, start, end uint64) (bytes.HexBytes, error) {
	client := s.getRPCClient()

	res, err := client.BridgeCommitment(ctx, start, end)
	if err != nil {
		return nil, err
	}
	return res.BridgeCommitment, nil
}

// Deprecated: Use QueryBridgeCommitmentInclusionProof instead.
func (s *E2ETestSuite) GetBridgeCommitmentInclusionProof(
	ctx context.Context, height, txIndex int64, start, end uint64,
) (*coretypes.ResultBridgeCommitmentInclusionProof, error) {
	client := s.getRPCClient()

	res, err := client.BridgeCommitmentInclusionProof(ctx, height, txIndex, start, end)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetBlockByHeight fetches the block at a given height. Note: we are explicitly using the res.Block type which has been
// deprecated instead of res.SdkBlock to support backwards compatibility tests.
func (s *E2ETestSuite) GetBlockByHeight(ctx context.Context, height int64) (*cmtservice.Block, error) {
	tmService := s.getGRPCClients().ConsensusServiceClient
	res, err := tmService.GetBlockByHeight(ctx, &cmtservice.GetBlockByHeightRequest{
		Height: height,
	})
	if err != nil {
		return nil, err
	}

	return res.SdkBlock, nil
}

func (s *E2ETestSuite) GetBlockResultsByHeight(ctx context.Context, height int64) (*coretypes.ResultBlockResults, error) {
	client := s.getRPCClient()
	return client.BlockResults(ctx, &height)
}

// GetValidatorSetByHeight returns the validators of the given Chain at the specified height. The returned validators
// are sorted by address.
func (s *E2ETestSuite) GetValidatorSetByHeight(ctx context.Context, height uint64) ([]*cmtservice.Validator, error) {
	tmService := s.getGRPCClients().ConsensusServiceClient
	res, err := tmService.GetValidatorSetByHeight(ctx, &cmtservice.GetValidatorSetByHeightRequest{
		Height: int64(height),
	})
	if err != nil {
		return nil, err
	}

	sort.SliceStable(res.Validators, func(i, j int) bool {
		return res.Validators[i].Address < res.Validators[j].Address
	})

	return res.Validators, nil
}
